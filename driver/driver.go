package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"syscall"

	mongodb "mongo-log-driver/mongo"

	"github.com/containerd/fifo"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type Driver struct {
	mu   sync.Mutex
	logs map[string]*LogPair
}

type LogPair struct {
	active     bool
	file       string
	info       logger.Info
	logLine    jsonLogLine
	stream     io.ReadCloser
	mongo      *mongo.Client
	collection *mongo.Collection
}

type MongoDriver struct {
	client     *mongo.Client
	collection *mongo.Collection
	// info       logger.Info
}

type MongoConf struct {
	uri        string
	dbname     string
	collection string
}

func NewDriver() *Driver {
	return &Driver{
		logs: make(map[string]*LogPair),
	}
}

func (d *Driver) StartLogging(file string, logCtx logger.Info) error {
	println("Start Logging")
	d.mu.Lock()
	if _, exists := d.logs[path.Base(file)]; exists {
		d.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	d.mu.Unlock()

	println("Checking Fifo")
	logrus.WithField("id", logCtx.ContainerID).WithField("file", file).Info("Start logging")
	stream, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	println("Parsing Log Tag")
	tag, err := loggerutils.ParseLogTag(logCtx, loggerutils.DefaultTemplate)
	if err != nil {
		return err
	}

	println("Extra Attributes")
	extra, err := logCtx.ExtraAttributes(nil)
	if err != nil {
		return err
	}

	println("Hostname")
	hostname, err := logCtx.Hostname()
	if err != nil {
		return err
	}

	logLine := jsonLogLine{
		ContainerId:      logCtx.FullID(),
		ContainerName:    logCtx.Name(),
		ContainerCreated: jsonTime{logCtx.ContainerCreated},
		ImageId:          logCtx.ImageFullID(),
		ImageName:        logCtx.ImageName(),
		Command:          logCtx.Command(),
		Tag:              tag,
		Extra:            extra,
		Host:             hostname,
	}

	println("Build mongo connection")
	client, collection, err := buildMongo(&logCtx, extra)
	if err != nil {
		return err
	}

	lp := &LogPair{true, file, logCtx, logLine, stream, client, collection}

	d.mu.Lock()
	d.logs[path.Base(file)] = lp
	d.mu.Unlock()

	// Call functino using logical thread -> go
	println("Call consume log as go")
	go consumeLog(lp)
	return nil
}

func (d *Driver) StopLogging(file string) error {
	logrus.WithField("file", file).Info("Stop logging")
	d.mu.Lock()
	lp, ok := d.logs[path.Base(file)]
	if ok {
		lp.active = false
		delete(d.logs, path.Base(file))
	} else {
		logrus.WithField("file", file).Errorf("Failed to stop logging. File %q is not active", file)
	}
	d.mu.Unlock()
	return nil
}

func shutdownLogPair(lp *LogPair) {
	if lp.stream != nil {
		lp.stream.Close()
	}

	// Close mongo connection
	if lp.mongo != nil {
		lp.mongo.Disconnect(context.Background())
	}

	lp.active = false
}

func consumeLog(lp *LogPair) {
	var buf logdriver.LogEntry

	dec := protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer shutdownLogPair(lp)

	for {
		if !lp.active {
			logrus.WithField("id", lp.info.ContainerID).Debug("shutting down logger goroutine due to stop request")
			return
		}

		err := dec.ReadMsg(&buf)
		if err != nil {
			if err == io.EOF {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Debug("shutting down logger goroutine due to file EOF")
				return
			} else {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error reading from FIFO, trying to continue")
				dec = protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
				continue
			}
		}

		err = logMessageToServer(lp, buf.Line)
		if err != nil {
			logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error logging message, dropping it and continuing")
		}

		buf.Reset()
	}
}

func buildMongo(logCtx *logger.Info, extra map[string]string) (*mongo.Client, *mongo.Collection, error) {
	println("Build Mongo Called!!")
	// Build mongo connection

	// Get config
	useOpt := readWithDefault(logCtx.Config, "use-opt", "false")
	// if env is true, use Environmental Variable.

	var server string
	var test string
	var dbname string
	var collection string
	println(useOpt)
	if useOpt == "false" {
		server = os.Getenv("LOG_MONGO_URL") // Get ENV from plugin set (GLOBAL)
		dbname = os.Getenv("LOG_MONGO_DBNAME")
		collection = os.Getenv("LOG_MONGO_COLLECTION")
	} else {
		server = readWithDefault(logCtx.Config, "server", "mongodb://0.0.0.0:27017")
		dbname = readWithDefault(logCtx.Config, "dbname", "docker-logs")
		collection = readWithDefault(logCtx.Config, "collection", "logs")
	}
	println(test)
	println(server)
	// Create MongoDB Connection
	client, err := mongodb.CreateMongoSession(server)
	if err != nil {
		fmt.Println("Failed to create Mongo Session", err, server)
		return nil, nil, err
	}

	// Check Connection
	err = mongodb.CheckMongoConnection(client)
	if err != nil {
		fmt.Println("Failed to connect to Mongo. Error: ", err)
	}

	// Get Collection
	coll := client.Database(dbname).Collection(collection)

	return client, coll, err

}

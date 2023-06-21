package driver

import (
	"encoding/json"
	"fmt"
	mongodb "mongo-log-driver/mongo"
	"time"
)

type jsonTime struct {
	time.Time
}

type jsonLogLine struct {
	Message          string            `json:"message"`
	ContainerId      string            `json:"container_id"`
	ContainerName    string            `json:"container_name"`
	ContainerCreated jsonTime          `json:"container_created"`
	ImageId          string            `json:"image_id"`
	ImageName        string            `json:"image_name"`
	Command          string            `json:"command"`
	Tag              string            `json:"tag"`
	Extra            map[string]string `json:"extra"`
	Host             string            `json:"host"`
	Timestamp        jsonTime          `json:"timestamp"`
}

func logMessageToServer(lp *LogPair, message []byte) error {
	lp.logLine.Message = string(message[:])
	lp.logLine.Timestamp = jsonTime{time.Now()}

	bytes, err := json.Marshal(lp.logLine)
	if err != nil {
		return err
	}
	println(string(bytes))
	return mongodb.InsertLogLine(bytes, lp.collection)
}

func (t jsonTime) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", t.Format(time.RFC3339Nano))
	return []byte(str), nil
}

// type LogDocument struct {
// 	ID          string `bson:"_id"`
// 	Timestamp   string `bson:"timestamp"`
// 	Message     string `bson:"message"`
// 	Container   string `bson:"container"`
// 	ContainerID string `bson:"container_id"`
// 	Image       string `bson:"image"`
// 	ImageID     string `bson:"image_id"`
// 	Host        string `bson:"host"`
// 	HostID      string `bson:"host_id"`
// 	Driver      string `bson:"driver"`
// 	DriverID    string `bson:"driver_id"`
// 	Source      string `bson:"source"`
// 	DriverName  string `bson:"driver_name"`
// 	DriverType  string `bson:"driver_type"`
// 	SourceID    string `bson:"source_id"`
// 	SourceName  string `bson:"source_name"`
// 	SourceType  string `bson:"source_type"`
// 	Level       string `bson:"level"`
// 	Tags        string `bson:"tags"`
// 	Meta        string `bson:"meta"`
// 	Fields      string `bson:"fields"`
// }

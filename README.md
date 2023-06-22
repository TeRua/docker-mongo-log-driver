# Docker MongoDB Log Driver

* Docker Registry
  * https://hub.docker.com/r/terua05/mongo-log-driver

## Purpose

This repository is Docker Log Driver to store logs into mongoDB directly not through Logstash or Fluentd to reduce dependecies on other services.

## Features

- Docker container logs store into MongoDB.
- Set MongoDB configs by environmental variables.
  - Mongo DB URL
  - Database Name
  - Collection name (global)
- Provide docker composer configuration
  - Configure for env variable usage
  - Collection name (per a service/container)
- Log format customize (TBD)
- Connection close at the end of program (including error, exception)

# Usage

## Build

To build docker plugin from the source, use build script
Build on Windows host is not supported.

`$ ./build.sh`

* in case you are stuck on plugin creation, try `sudo ./build.sh` instead.

## Pull from DockerHub

To pull plugin from DockerHub

`$ docker plugin pull terua05/mongo-log-driver:latest`

## Plugin's log

- To see plugin's logs, plugin must be used by root account (Use `su`)
- Plugin's socket directory will be created in `/run/docker/plugins/${PLUGIN_ID}`
  - Monitor stderr and stdout to see plugin's log

## How to use

### With Docker

- Use `--log-driver` option in Docker CLI
  - `docker run -d --log-driver mongo-log-driver:0.0.1 --log-opt use-opt=true --log-opt server="${SERVER_URI}" --log-opt dbname="${DB_NAME}" --log-opt collection="${COLLECTION_NAME}" terua05/mongo-log-driver `

### Docker Composer

- Use  `logging` key to configure logging driver in docker-compose

```yaml
version: '3.8'
services:
  test-composer:
    image: {image_name}
    ports:
      - 9000:9000
    logging:
        driver: terua/mongo-log-driver
        options:
          log-opt: "true"
          server: "mongodb://0.0.0.0:27017"
	  dbname: "test-db"
	  collection: "collection"
    networks:
      - log-test

networks:
  log-test:
    external: true
```

### How to set Environmental Variable

1. To set plugin's environmental variables, we need to use `docker plugin set ${ENV_VAR}=${VALUE}`.
   * Make sure the plugin is disabled before changing the env values.
     `docker plugin disable mongo-log-driver`
2. Only pre-defined environmental variables are able to configured. if not, it will b e ignored.
3. To check values assigned, use `docker plugin inspect mongo-log-driver`

## Variables

### Environmental Variable

1. LOG_MONGO_URL : MongoDB server address. following mongoDB documentation. (Default: mongodb://localhost:27017)
   - For MongoDB Atlas (Cloud) : `mongodb+srv://[username:password@]host1[,...hostN][/[defaultauthdb][?options]]`
     - `mongodb+srv` protocol will return error if port is described in the URI
   - For Self-hosted : `mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]`
2. LOG_MONGO_DBNAME : Name of database (Default : docker-logs)
3. LOG_MONGO_COLLECTION : Name of collection (Default : logs)

### Docker Variable (CLI or Docker-compose)

1. use-opt : You must put "true" for this option to use docker log option not environmental variables. (Default : false)
2. server : MongoDB server address. following mongoDB documentation.
   - For MongoDB Atlas (Cloud) : `mongodb+srv://[username:password@]host1[,...hostN][/[defaultauthdb][?options]]`
     - `mongodb+srv` protocol will return error if port is described in the URI
   - For Self-hosted : `mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]`
3. dbname : Name of database (Default : docker-logs)
4. collection : Name of collection (Default : logs)

# Appendix

## Reference

1. pressrelations/docker-redis-log-driver
   * [pressrelations/docker-redis-log-driver: Redis log driver for Docker (github.com)](https://github.com/pressrelations/docker-redis-log-driver)
   * https://hub.docker.com/r/pressrelations/docker-redis-log-driver
2. MongoDB URI syntax
   * https://www.mongodb.com/docs/manual/reference/connection-string/
3. Docker Plugin Doc
   * https://docker-docs.uclv.cu/engine/reference/commandline/plugin_set/

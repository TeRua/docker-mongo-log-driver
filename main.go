package main

import (
	"mongo-log-driver/driver"
	"mongo-log-driver/handler"

	"github.com/docker/go-plugins-helpers/sdk"
)

func main() {

	h := sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`)
	handler.Handlers(&h, driver.NewDriver())
	if err := h.ServeUnix("log", 0); err != nil {
		panic(err)
	}
}

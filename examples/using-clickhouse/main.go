package main

import (
	"gofr.dev/examples/using-clickhouse/handler"
	"gofr.dev/examples/using-clickhouse/store"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.New()

	s := store.New()
	h := handler.New(s)

	// specifying the different routes supported by this service
	app.GET("/user", h.Get)
	app.POST("/user", h.Create)
	app.GET("/user/{id}", h.GetByID)

	app.Server.HTTP.Port = 9002
	app.Server.MetricsPort = 2113

	app.Start()
}
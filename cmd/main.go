package main

import (
	"context"

	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/service"
	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	service.Main(ctx, cancel, mux.NewRouter())
}

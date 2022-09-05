package service

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/adapter"
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/config"
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/container"
	"github.com/edgexfoundry/edgexfoundry-iotdb-connector/internal/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"
)

var (
	AdapterServiceKey = "iotdb-adapter-service"
	ConfigStemCore    = ""
)

func Main(ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(AdapterServiceKey)

	f := flags.New()
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	edgeXAdapter := adapter.NewAdaptor()
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		adapter.EdgeXAdapterName: func(get di.Get) interface{} {
			return edgeXAdapter
		},
	})

	httpServer := handlers.NewHttpServer(router, true)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		AdapterServiceKey,
		ConfigStemCore,
		configuration,
		startupTimer,
		dic,
		false,
		[]interfaces.BootstrapHandler{
			httpServer.BootstrapHandler,
			edgeXAdapter.BootstrapHandler,
			messaging.BootstrapHandler,
			messaging.NewBootstrap().BootstrapHandler,
			handlers.NewStartMessage(AdapterServiceKey, "0.1.0").BootstrapHandler,
		})
}

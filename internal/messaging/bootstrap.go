package messaging

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router      *mux.Router
	serviceName string
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
// func NewBootstrap(router *mux.Router, serviceName string) *Bootstrap {
// 	return &Bootstrap{
// 		router:      router,
// 		serviceName: serviceName,
// 	}
// }

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {

	lc := container.LoggingClientFrom(dic.Get)

	err := SubscribeEvents(ctx, dic)
	if err != nil {
		lc.Errorf("Failed to subscribe events from message bus, %v", err)
		return false
	}

	return true
}

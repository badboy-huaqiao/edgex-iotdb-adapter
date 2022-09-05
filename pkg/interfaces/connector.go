package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

type Connector interface {
	EventWriter(dtos.Event) error
	ReadingWriter(dtos.BaseReading) error
	TsFileSync() error
}

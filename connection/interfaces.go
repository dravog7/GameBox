package connection

import (
	"time"

	"github.com/google/uuid"
)

type Sender interface {
	Send(string) error
	SendJSON(interface{}) error
}

type Receiver interface {
	Subscribe(chan *InMessage) uuid.UUID
	Unsubscribe(uuid.UUID) error
}

type Pinger interface {
	GetPing() time.Duration
}

type Closer interface {
	Close() error
	IsClosed() bool
}

type UUIDable interface {
	GetUUID() uuid.UUID
}

type IConnection interface {
	Sender
	Receiver
	Closer
	Pinger
	UUIDable
}

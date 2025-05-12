package transport

import (
	"context"
	"io"

	"github.com/mohammadhptp/pulse/pkg/models"
)

type EventProducer interface {
	Start(ctx context.Context) error
	Stop() error
	SetEventHandler(handler EventHandler)
	io.Closer
}

type EventHandler func(event models.Event) error

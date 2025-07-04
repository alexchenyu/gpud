package events

import (
	"context"
	"time"

	"github.com/leptonai/gpud/pkg/eventstore"
	"github.com/leptonai/gpud/pkg/nvidia-query/infiniband"
)

// IBPortsStore defines the interface for storing IB ports events.
type IBPortsStore interface {
	// Insert inserts the IB ports into the store.
	// The timestamp is the time when the IB ports were queried.
	// Only stores the "Infiniband" link layer ports (not "Ethernet" or "Unknown").
	Insert(ctx context.Context, event *IBPortsEvent) error
	// Get returns the all IB ports events since the given time.
	Get(ctx context.Context, since time.Time) (IBPortsEvents, error)
}

// IBPortsEvent represents an IB ports event,
// which contains the time and the IB ports.
type IBPortsEvent struct {
	Time    time.Time
	IBPorts []infiniband.IBPort
}

// IBPortsEvents is a slice of IB ports events.
type IBPortsEvents []IBPortsEvent

var _ IBPortsStore = &ibPortsStore{}

type ibPortsStore struct {
	eventsStore eventstore.Store
}

func NewIBPortsStore(eventsStore eventstore.Store) IBPortsStore {
	return &ibPortsStore{
		eventsStore: eventsStore,
	}
}

func (s *ibPortsStore) Insert(ctx context.Context, event *IBPortsEvent) error {
	return nil
}

func (s *ibPortsStore) Get(ctx context.Context, since time.Time) (IBPortsEvents, error) {
	return nil, nil
}

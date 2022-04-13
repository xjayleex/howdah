package howdah_agent

import (
	"context"
	"google.golang.org/grpc"
	"howdah/internal/pkg/howdah-agent"
	"howdah/pb"
)

type EventWatcher struct {
	eventClient   pb.HowdahEventClient
	eventProducer howdah_agent.EventProducer
}

func NewEventWatcher(cc grpc.ClientConnInterface, eventProducer howdah_agent.EventProducer) *EventWatcher {

	return &EventWatcher{
		eventClient: pb.NewHowdahEventClient(cc),
		eventProducer: eventProducer,
	}
}

func (w *EventWatcher) Run(ctx context.Context) error {
	recovery := func() {
		if err := recover(); err != nil {
			// TODO Inject recovery policy.
		}
	}
	defer recovery()

	// Available error ?
	// 1. TCP Connection (Server down, ...)
	stream, err := w.eventClient.Watch(
		// Todo Verification must be originated from registration.
		ctx, &pb.Verification{Ok: true},
	)

	if err != nil {
		panic(err)
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		w.pushEvent(event)
	}
}

func (w *EventWatcher) pushEvent(event *pb.Event) {
	w.eventProducer.Produce(event)
}

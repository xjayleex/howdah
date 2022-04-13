package howdah_agent

import (
	"context"
	"errors"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/sirupsen/logrus"
	howdah_agent "howdah/internal/pkg/howdah-agent/grpc"
	"howdah/pb"
	"time"
)

type EventProducer interface {
	Produce(*pb.Event) error
}

type eventProducer struct {
	eventQueue EventQueue
}

func NewEventProducer(queue EventQueue) *eventProducer {
	return &eventProducer{
		eventQueue: queue,
	}
}

func (ep *eventProducer) Produce(event *pb.Event) error {
	// ep.eventQueue.Add()
	return nil
}

type EventRoutine struct {
	logger 		*logrus.Logger
	eventWatcher	*howdah_agent.EventWatcher
	eventProcessor EventProcessor
}

func (er *EventRoutine) Run () {
	ctx := context.Background()
	er.eventWatcher.Run(ctx)
}

type EventProcessor struct {
	eventConsumer EventConsumer
}

func NewEventProcessor (eventConsumer EventConsumer) *EventProcessor {
	return &EventProcessor{
		eventConsumer: eventConsumer,
	}
}

func (ep *EventProcessor) Run () {
	// TODO Increase wait-group and defer decrease.
	for {
		event, err := ep.eventConsumer.Consume()
		if err != nil {
			// TODO What to do with these error?
			Global().DebugLogger().Log(logrus.DebugLevel, "error on consuming ... err\n%v", err)
			continue
		}
		ep.handle(event)
	}
}

func (ep *EventProcessor) handle(event *pb.Event) {
	// TODO
}

type EventConsumer interface {
	Consume() (*pb.Event, error)
}

type eventConsumer struct {
	eventQueue EventQueue
}

func NewEventConsumer(queue EventQueue) *eventConsumer {
	return &eventConsumer{
		eventQueue: queue,
	}
}

func (ec *eventConsumer) Consume() (*pb.Event, error) {
	event, err := ec.eventQueue.Poll()
	return event, err
}

type EventQueue interface {
	Add(event *pb.Event) error
	Poll() (*pb.Event, error)
	PollWithTimeout(duration time.Duration) (*pb.Event, error)
	// Pop() (*pb.Event, error)
}


type eventQueue struct {
	queue goconcurrentqueue.Queue
}

func NewEventQueue (queue goconcurrentqueue.Queue) *eventQueue {
	return &eventQueue{
		queue: queue,
	}
}

func (eq *eventQueue) Add(event *pb.Event) error {
	for !eq.queue.IsLocked(){
		// Wait until getting lock.
	}
	err := eq.queue.Enqueue(event)
	return err
}

func (eq *eventQueue) Poll() (*pb.Event, error) {
	elem, err := eq.queue.DequeueOrWaitForNextElement()
	if err != nil {
		return nil, err
	}
	if event, ok := elem.(*pb.Event); !ok {
		return nil, errors.New("type assertion error on pb.Event type")
	} else {
		return event, nil
	}
}

func (eq *eventQueue) PollWithTimeout(timeout time.Duration) (*pb.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	elem, err := eq.queue.DequeueOrWaitForNextElementContext(ctx)
	if err != nil {
		return nil, err
	}
	if event, ok := elem.(*pb.Event); !ok {
		return nil, errors.New("type assertion error on pb.Event type")
	} else {
		return event, nil
	}
}

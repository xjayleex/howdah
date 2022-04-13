package main

import (
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	howdah_agent "howdah/internal/pkg/howdah-agent"
	howdah_grpc "howdah/internal/pkg/howdah-agent/grpc"
	"sync"
	"time"
)


func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	addr := "localhost:9091"
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatal(err)
	}
	heartbeatClient := howdah_grpc.NewHeartbeatReceptionClient(conn)
	// EventWathcer
	queue := goconcurrentqueue.NewFIFO()
	eventQueue := howdah_agent.NewEventQueue(queue)
	eventProducer := howdah_agent.NewEventProducer(eventQueue)
	eventWatcher := howdah_grpc.NewEventWatcher(conn, eventProducer)

	hbRoutine, err := howdah_agent.NewHeartbeatRoutine(logger,
		heartbeatClient,
		eventWatcher,
		howdah_agent.WithHeartbeatInterval(time.Second * 10),
		howdah_agent.WithHeartbeatTimeout(time.Second * 30),
	)
	if err != nil {
		logrus.Fatalln()
	}
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(3)
	go func () {
		defer wg.Done()
		hbRoutine.Run()
	}()
	go func () {
		defer wg.Done()
		hbRoutine.Watch()
	}()


}

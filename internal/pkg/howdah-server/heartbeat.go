package howdah_server

import (
	"context"
	"github.com/enriquebris/goconcurrentqueue"
	"howdah/internal/pkg/common/utils"
	"howdah/pb"
)


type HeartbeatHandler interface {
	HandleHeartbeat(ctx context.Context, heartbeat *pb.Heartbeat) (*pb.HeartbeatResponse, error)
}

type heartbeatHandler struct {
	// logger
	// clusterFsm Clusters
	// Encrypter
	// HeartbeatMonitor
	// config Configuration
	// ambariMetaInfo AmbariMetaInfo
	timestamper *utils.Timestamper
	processor HeartbeatProcessor

}

func NewHeartbeatHandler (processor HeartbeatProcessor, timestamper *utils.Timestamper) *heartbeatHandler {
	return &heartbeatHandler{
		processor: processor,
		timestamper: timestamper,
	}
}

func (hh *heartbeatHandler) HandleHeartbeat (ctx context.Context, heartbeat *pb.Heartbeat) (*pb.HeartbeatResponse, error){
	hh.processor.addHeartbeat(heartbeat)
	return &pb.HeartbeatResponse{
		ResponseId:              0,
		ExecutionCommands:       nil,
		StatusCommands:          nil,
		CancelCommands:          nil,
		AlertDefinitionCommands: nil,
		RegistrationCommand:     nil,
		RestartAgent:            false,
		HasMappedComponents:     false,
		HasPendingTasks:         false,
		RecoveryConfig:          nil,
		ClusterSize:             0,
	}, nil
}

type HeartbeatProcessor interface {
	Start()
	Stop()

	addHeartbeat(heartbeat *pb.Heartbeat)
	pollHeartbeat() *pb.Heartbeat
}

type heartbeatProcessor struct {
	heartbeatQueue HeartbeatQueue
}

func NewHeartbeatProcessor(queue HeartbeatQueue) *heartbeatProcessor {
	return &heartbeatProcessor{heartbeatQueue: queue}
}

func (hp *heartbeatProcessor) Start() {}
func (hp *heartbeatProcessor) Stop() {}
func (hp *heartbeatProcessor) addHeartbeat(heartbeat *pb.Heartbeat) {
	hp.heartbeatQueue.Add(heartbeat)
}
func (hp *heartbeatProcessor) pollHeartbeat() *pb.Heartbeat {
	return hp.heartbeatQueue.Poll()
}

type HeartbeatQueue interface {
	Add(heartbeat *pb.Heartbeat)
	Poll() *pb.Heartbeat
}

type concurrrentQueue struct {
	queue goconcurrentqueue.Queue
}

func NewConcurrentQueue (queue goconcurrentqueue.Queue) *concurrrentQueue {
	return &concurrrentQueue{queue: queue}
}

func (cq *concurrrentQueue) Add(heartbeat *pb.Heartbeat) {
}

func (cq *concurrrentQueue) Poll() *pb.Heartbeat {
	return nil
}

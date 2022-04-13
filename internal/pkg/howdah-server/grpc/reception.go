package howdah_server

import (
	"context"
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	howdah_server "howdah/internal/pkg/howdah-server"
	"howdah/pb"
	"time"
)

type HeartbeatReceptionServer struct {
	logger *logrus.Logger
	receptionist Receptionist
	heartbeatHandler howdah_server.HeartbeatHandler
}

func NewHeartbeatReceptionServer(logger *logrus.Logger, receptionist Receptionist, heartbeatHandler howdah_server.HeartbeatHandler) *HeartbeatReceptionServer {

	rs := &HeartbeatReceptionServer{
		logger: logger,
		receptionist: receptionist,
		heartbeatHandler: heartbeatHandler,
	}
	return rs
}

func (rs *HeartbeatReceptionServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (resp *pb.RegisterAgentResponse, err error) {
	rs.logger.Log(logrus.DebugLevel, "Registration called.")
	if err := rs.receptionist.HandleRegistration(ctx, req); err != nil {
		rs.logger.Log(logrus.InfoLevel, err)
		return nil, err
	}
	return &pb.RegisterAgentResponse{
		Ok: true,
	}, nil
}

func (rs *HeartbeatReceptionServer) AgentHeartbeat(ctx context.Context, req *pb.Heartbeat) (*pb.HeartbeatResponse, error) {
	// Forcing timeout.
	// time.Sleep(time.Second * 10)
	return rs.heartbeatHandler.HandleHeartbeat(ctx, req)
}

type Receptionist interface {
	HandleRegistration(context.Context, *pb.RegisterAgentRequest) error
}

type AgentInfo struct {
	fqdn string
	ipaddr string
	active bool
}

type AgentInfoStore interface {
	SaveAgentInfo(info AgentInfo) error
}


type HowdahEventServer struct {
	eventHandler *HowdahEventHandler
}

func NewHowdahEventServer (eventHandler *HowdahEventHandler) *HowdahEventServer {
	return &HowdahEventServer{eventHandler: eventHandler}
}


func (s *HowdahEventServer) Watch(in *pb.Verification, stream pb.HowdahEvent_WatchServer) error {
	/* Todo belows.
	1. Run verification for the registration
	2. Serve streams.
	 */
	stream.Send(&pb.Event{
		EventType: &pb.Event_AgentAction{
			AgentAction: &pb.Event_AgentActionEvent{Type: "Stream started"},
		},
	})

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Minute)
		defer cancel()
		event, err := s.eventHandler.Poll(ctx)
		if err != nil {
			fmt.Println(err)
		}
		stream.Send(event)

	}

	return status.Errorf(codes.Unimplemented, "method Watch not implemented") // FIXME
}


type HowdahEventHandler struct {
	queue goconcurrentqueue.Queue
}

func NewHowdahEventHandler (queue goconcurrentqueue.Queue) *HowdahEventHandler{
	return &HowdahEventHandler{
		queue: queue,
	}
}

func (eh *HowdahEventHandler) Poll (ctx context.Context) (*pb.Event, error) {
	v, err := eh.queue.DequeueOrWaitForNextElementContext(ctx)
	if err != nil {
		return nil, err
	}
	event := v.(*pb.Event)
	return event, nil
}

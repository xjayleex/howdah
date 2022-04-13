package howdah_agent

import (
	"context"
	"google.golang.org/grpc"
	"howdah/pb"
)


type HeartbeatReceptionClient struct {
	client pb.HeartbeatReceptionClient
	// Not thread safe yet.
}

func NewHeartbeatReceptionClient(cc grpc.ClientConnInterface) *HeartbeatReceptionClient {
	return &HeartbeatReceptionClient{
		client:      pb.NewHeartbeatReceptionClient(cc),
	}
}

func (c *HeartbeatReceptionClient) RegisterAgent(ctx context.Context, in *pb.RegisterAgentRequest, opts ...grpc.CallOption) (*pb.RegisterAgentResponse, error) {
	return c.client.RegisterAgent(ctx, in, opts...)
}

func (c *HeartbeatReceptionClient) Heartbeat(ctx context.Context, in *pb.Heartbeat, opts ...grpc.CallOption) (*pb.HeartbeatResponse, error ){
	return c.client.AgentHeartbeat(ctx, in, opts...)
}

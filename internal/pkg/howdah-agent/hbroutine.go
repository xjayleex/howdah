package howdah_agent

import (
	"context"
	"github.com/sirupsen/logrus"
	"howdah/internal/pkg/common/const"
	"howdah/internal/pkg/common/infra"
	"howdah/internal/pkg/common/utils"
	howdah_agent "howdah/internal/pkg/howdah-agent/grpc"
	"howdah/pb"
	"time"
)

//
type HeartbeatRoutine struct {
	logger            *logrus.Logger
	template          *RegisterTemplate
	generator         *heartbeatGenerator

	opts heartbeatOptions

	shouldStop        *bool
	heartbeatClient   *howdah_agent.HeartbeatReceptionClient

}

func NewHeartbeatRoutine (logger *logrus.Logger, heartbeatClient *howdah_agent.HeartbeatReceptionClient, eventClient *howdah_agent.EventWatcher, opt ...HeartbeatOption) (*HeartbeatRoutine, error) {
	register, err := NewRegisterTemplate()

	if err != nil {
		return nil, err
	}

	generator, err := NewHeartbeatGenerator()

	if err != nil {
		return nil, err
	}

	opts := defaultHeartbeatOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	return &HeartbeatRoutine{
		logger:            logger,
		template:          register,
		generator:		   generator,
		heartbeatClient:   heartbeatClient,
		// FIXME : where would shouldStop set?
		opts: opts,
		shouldStop: &Global().shouldStop,
	}, nil


}

func (hr *HeartbeatRoutine) Run() {
	for !(*hr.shouldStop) {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(),
				hr.opts.heartbeatTimeout)
			defer cancel()
			err := hr.heartbeats(ctx)
			if err != nil {
				hr.logger.Log(logrus.InfoLevel, err)
				if err == context.DeadlineExceeded {
					hr.logger.Log(logrus.InfoLevel, "Cancled.")
				}
			}
		}()

		time.Sleep(hr.opts.heartbeatInterval)
	}
}

func (hr *HeartbeatRoutine) heartbeats(ctx context.Context) error {
	recovery := func() {
		if err := recover(); err != nil {
				Global().SetRegistered(false)
			}
		}

	defer recovery()

	if !Global().Registered() {
		// FIXME : try to register
		if ok := hr.register(); !ok {
			panic("Connection error. Re-running the registration.")
		}
	}

	hr.logger.Log(logrus.DebugLevel,"Heartbeating...\n")

	heartbeat := hr.generator.Generate()
	if heartbeat == nil {
		hr.logger.Log(logrus.DebugLevel, "heartbeat object is nil")
	}
	_, err := hr.heartbeatClient.Heartbeat(ctx, heartbeat)
	if err != nil {
		hr.logger.Log(logrus.DebugLevel, "error on sending", err)
		panic(err)
	}
	return nil
}

func (hr *HeartbeatRoutine) register() bool {
	// FIXME : hr.heartbeatClient should implement this logic.
	resp, err := hr.heartbeatClient.RegisterAgent(
				context.Background(),
				hr.template.Build())

	if err != nil {
		// FIXME : HandleRegistration gRPC request error.
		hr.logger.Logln(logrus.InfoLevel, err)
		return false
	}

	return Global().SetRegistered(resp.Ok)
}

type heartbeatGenerator struct {
	hostname string
	timestamper *utils.Timestamper
}

func NewHeartbeatGenerator () (*heartbeatGenerator, error){
	hostname, err := infra.Hostname()
	if err != nil {
		return nil, err
	}
	timestamper := utils.NewTimestamper()
	return &heartbeatGenerator{
		hostname: hostname,
		timestamper: timestamper,
	}, nil
}


func (hg *heartbeatGenerator) Generate() *pb.Heartbeat {
	heartbeat := &pb.Heartbeat{
		Timestamp: hg.timestamper.Now(),
		HostName: hg.hostname,
		NodeStatus: &pb.HostStatus{
			Status: 0,
			Cause:  "",
		},
		ComponentStatus: nil,
		Alerts:          nil,
		AgentEnv: &pb.AgentEnv{
			StackFoldersAndFiles: nil,
			Alternatives:         nil,
			ExistingUsers:        nil,
			ExistingRepos:        nil,
			InstalledPackages:    nil,
			HostHealth: &pb.AgentEnv_HostHealth{
				AciveJavaProcs:             nil,
				AgentTimeStampAtReporting:  0,
				ServerTimeStampAtReporting: 0,
				LiveServices:               nil,
			},
			Umask:                 0,
			TransparentHugePage:   "",
			FirewallRunning:       false,
			FirewallName:          "",
			HasUnlimitedJcePolicy: false,
			ReverseLookup:         false,
		},
	}
	return heartbeat
}

type HeartbeatOption interface {
	apply(*heartbeatOptions)
}

type heartbeatOptions struct {
	heartbeatInterval time.Duration
	heartbeatTimeout time.Duration
}
var defaultHeartbeatOptions = heartbeatOptions {
	heartbeatInterval: consts.HeartbeatInterval,
	heartbeatTimeout:  consts.HeartbeatTimeout,
}

func WithHeartbeatInterval (interval time.Duration) HeartbeatOption{
	return newFuncHeartbeatOption(func(o *heartbeatOptions) {
		o.heartbeatInterval = interval
	})
}

func WithHeartbeatTimeout (timeout time.Duration) HeartbeatOption {
	return newFuncHeartbeatOption(func(o *heartbeatOptions) {
		o.heartbeatTimeout = timeout
	})
}


type funcHeartbeatOption struct {
	f func(*heartbeatOptions)
}

func (fho *funcHeartbeatOption) apply(to *heartbeatOptions) {
	fho.f(to)
}

func newFuncHeartbeatOption(f func(*heartbeatOptions)) *funcHeartbeatOption {
	return &funcHeartbeatOption{
		f: f,
	}
}

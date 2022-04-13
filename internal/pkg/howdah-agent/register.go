package howdah_agent

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"howdah/internal/pkg/common/infra"
	"howdah/internal/pkg/common/utils"
	"howdah/pb"
)

type RegisterTemplate struct {
	timestamper *utils.Timestamper
	// Todo : We can replace fields below used for setting RegisterAgentRequest with a template struct.
	agentStartTime *timestamppb.Timestamp
	hostname string
}

func NewRegisterTemplate() (*RegisterTemplate, error) {
	timestamper := utils.NewTimestamper()
	agentStartTime := timestamper.Now()
	hostname, err := infra.Hostname()

	if err != nil {
		return nil, err
	}

	return &RegisterTemplate{
		timestamper: timestamper,
		agentStartTime: agentStartTime,
		hostname: hostname,
	}, nil
}

func (r *RegisterTemplate) Build () *pb.RegisterAgentRequest {
	template := &pb.RegisterAgentRequest{
		ResponseId: -1,
		Timestamp: r.timestamper.Now(),
		AgentStartTime: r.agentStartTime,
		Hostname:        r.hostname,
		CurrentPingPort: 0,
		HardwareProfile: &pb.HostInfo{
			Architecture:           "",
			Domain:                 "",
			Fdqn:                   "",
			HardwareIsa:            "",
			HardwareModel:          "",
			Hostname:               "",
			Id:                     "",
			Interfaces:             "",
			IpAddress:              "",
			Kernel:                 "",
			KernelMajorVersion:     "",
			KernelRelease:          "",
			KernelVersion:          "",
			MacAddress:             "",
			MemoryFree:             0,
			MemorySize:             0,
			MemoryTotal:            0,
			Mounts:                 nil,
			Netmask:                "",
			OperatingSystem:        "",
			OperatingSystemRelease: "",
			OsFamily:               "",
			PhysicalProcessorCount: 0,
			ProcessorCount:         0,
			Selinux:                false,
			SwapFree:               "",
			SwapSize:               "",
			Timezone:               "",
			Uptime:                 "",
			UptimeDays:             0,
			UptimeHours:            0,
		},
		PublicHostName: "",
		AgentEnv:       &pb.AgentEnv{
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
		AgentVersion:   "",
		Prefix:         "",
	}

	return template
}
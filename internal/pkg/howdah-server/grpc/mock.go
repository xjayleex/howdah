package howdah_server

import (
	"context"
	"errors"
	"fmt"
	"howdah/pb"
)

type mockReceptionist struct {
	store AgentInfoStore
}

func NewMockReceptionist (store AgentInfoStore) mockReceptionist {
	return mockReceptionist{
		store: store,
	}
}

func (r *mockReceptionist) HandleRegistration(ctx context.Context, req *pb.RegisterAgentRequest) error {
	agentInfo := AgentInfo{
		fqdn:   req.Hostname,
		ipaddr: req.HardwareProfile.IpAddress,
		active: false,
	}
	err := r.store.SaveAgentInfo(agentInfo)

	return err
}

type mockAgentInfoStore map[string]AgentInfo

func NewMockAgentStore () mockAgentInfoStore {
	store := make(mockAgentInfoStore)
	return store
}

func (s mockAgentInfoStore) SaveAgentInfo(info AgentInfo) error {
	// Fixme To be deleted
	fmt.Println("SaveAgentInfo Called.")

	if _, exists := s[info.fqdn]; exists {
		// return status.Error(codes.AlreadyExists, "already registered host.")
		return nil
	}

	s[info.fqdn] = info

	return nil
}

type mockAdminStore map[string]admin

func NewMockAdminStore () mockAdminStore {
	admin := admin{
		id:       "admin",
		password: "admin",
	}

	store := make(mockAdminStore)
	store[admin.id] = admin
	return store
}

func (store mockAdminStore) find(id string) (admin, error) {
	if admin, exists := store[id]; !exists {
		return admin, errors.New("Given ID does not exist.")
	} else {
		return admin, nil
	}
}
package infra

import (
	"context"
	"github.com/contiv/executor"
	"github.com/sirupsen/logrus"
	"howdah/internal/pkg/common/codes"
	"howdah/internal/pkg/common/status"
	"os/exec"
	"time"
)

type OsCommandExecutor interface {
	Execute (ctx context.Context, cmd *exec.Cmd) (*executor.ExecResult, error)
	ExecuteWithTimeout(timeout time.Duration, cmd *exec.Cmd) (*executor.ExecResult, error)
	ExecuteWithExitOnHang(timeout time.Duration, cmd *exec.Cmd) (*executor.ExecResult, error)
}

type osCommandExecutor struct {
	logger *logrus.Logger
}

func NewOsCommandExecutor(logger *logrus.Logger) OsCommandExecutor {
	return &osCommandExecutor{
		logger: logger,
	}
}

func (ce *osCommandExecutor) Execute(ctx context.Context, cmd *exec.Cmd) (*executor.ExecResult, error) {
	exec := executor.NewCapture(cmd)
	err := exec.Start()
	if err != nil {
		return nil, err
	}

	result, err := exec.Wait(ctx)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (ce *osCommandExecutor) ExecuteWithTimeout(timeout time.Duration, cmd *exec.Cmd) (*executor.ExecResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ch := make(chan func() (*executor.ExecResult, error))
	go func(ctx context.Context, cmd *exec.Cmd) {
		result, err := ce.Execute(ctx, cmd)
		ch <- func()(*executor.ExecResult, error){
			return result, err
		}
	} (ctx, cmd)

	select {
	case <- ctx.Done():
		// Timeout.
		return nil, status.Errorf(codes.OsCommandTimeout, "")
	case resultChan := <-ch:
		result, err := resultChan()
		return result, err
	}
}
func (ce *osCommandExecutor) ExecuteWithExitOnHang(timeout time.Duration, cmd *exec.Cmd) (*executor.ExecResult, error) {
	return nil, status.Errorf(codes.NotImplemented, "")
}

package howdah_agent

import (
	"github.com/sirupsen/logrus"
	"howdah/internal/pkg/common/const"
	"sync"
)

var gb *global
var once sync.Once

type global struct {
	shouldStop bool
	registered bool
	debugLogger *logrus.Logger
}

func Global() *global {
	once.Do(
		func() {
			logger := logrus.New()
			logger.SetLevel(consts.LogLevel)
			gb = &global{
				shouldStop: false,
				registered	: false,
				debugLogger: logger,
			}
			// Can there be an error ?
	})
	return gb
}

func (g *global) Registered() bool {
	return g.registered
}

func (g *global) SetRegistered(r bool) bool {
	g.registered = r
	return g.registered
}

func (g *global) DebugLogger() *logrus.Logger {
	return g.debugLogger
}
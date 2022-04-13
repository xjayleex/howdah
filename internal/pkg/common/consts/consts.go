package consts

import (
	"time"
	"github.com/sirupsen/logrus"
)

const (
	LogLevel = logrus.DebugLevel
)

const (
	MinRecoveryInterval = 200 * time.Millisecond
	MaxRecoveryInterval = 3000 * time.Millisecond
)

const (
	HeartbeatInterval = 3 * 1000 * time.Millisecond
	HeartbeatTimeout = 1.5 * 1000 * time.Millisecond
)

const (
	DefaultTimezone = "Asia/Seoul"
)

// File Types
const (
	NotExists = -2 + iota
	Unknown  // includes irregular file.
	File
	Directory
	Symlink
)
package codes

import "strconv"

type Code uint32

func (c Code) String() string {
	switch c {
	case PackageAlreadyInstalled:
		return "PackageAlreadyInstalled"
	case PackageNotAvailable:
		return "PackageNotAvailable"
	case Unknown:
		return "Unknown"
	case UnknownOsFamily:
		return "UnknownOsFamily"
	case NotImplemented:
		return "NotImplemented"
	case InvalidArguments:
		return "InvalidArguments"
	case OsCommandTimeout:
		return "OsCommandTimeout"
	default:
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	}
}

const (
	PackageAlreadyInstalled = 0 + iota
	PackageNotAvailable
	Unknown
	UnknownOsFamily
	NotImplemented
	InvalidArguments
	OsCommandTimeout
)

package infra

type OsFamily int32

const (
	OsFamily_REDHAT  OsFamily = 0
	OsFamily_DEBIAN  OsFamily = 1
	OsFamily_UNKNOWN OsFamily = 2
)

func GetOsFamily() OsFamily {
	// Fixme : still Mocking...
	return OsFamily_REDHAT
}

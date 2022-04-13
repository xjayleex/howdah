package howdah_agent

import (
	"golang.org/x/sys/unix"
	"howdah/internal/pkg/common/const"
	"os"
)

type HostInfo interface{
	Register() error
}

type hostInfo struct {
	currentUmask int
}

// Equivalent with dirType.
func (hi *hostInfo) fileType (path string) int {
	stat, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return consts.NotExists
		} else {
			return consts.Unknown
		}
	}

	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		return consts.Symlink
	}
	if stat.Mode().IsDir() {
		return consts.Directory
	}
	if stat.Mode().IsRegular() {
		return consts.File
	}

	return consts.Unknown
}

func (hi *hostInfo) checkLiveServices(){}

// Umask(Oct & Dec)
// 		 000    0
//		 002	2
//		 022   18
//		 027   23
//		 077   63
func (hi *hostInfo) umask() int {
	if hi.currentUmask == -1 {
		hi.currentUmask = unix.Umask(hi.currentUmask)
		unix.Umask(hi.currentUmask)
		return hi.currentUmask
	}
	return hi.currentUmask
}


func DefaultHostInfo () hostInfo {
	return hostInfo{
		// TODO
		currentUmask: -1,
	}
}

type LinuxHostInfo struct {
	hostInfo
}

func NewLinuxHostInfo () *LinuxHostInfo {
	hostInfo := DefaultHostInfo()
	return &LinuxHostInfo{
		hostInfo: hostInfo,
	}
}

func (lhi *LinuxHostInfo) Register() error {
	return nil
}

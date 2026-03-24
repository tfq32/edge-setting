//go:build windows

package collector

type SysInfo struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func getSysInfo() (*SysInfo, error) {
	return &SysInfo{
		Sysname: "Windows",
	}, nil
}

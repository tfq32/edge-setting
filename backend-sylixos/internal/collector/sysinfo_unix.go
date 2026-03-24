//go:build !windows

package collector

import (
	"golang.org/x/sys/unix"
)

// SysInfo 通过 unix.Uname 获取系统信息
type SysInfo struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func getSysInfo() (*SysInfo, error) {
	var buf unix.Utsname
	if err := unix.Uname(&buf); err != nil {
		return nil, err
	}
	return &SysInfo{
		Sysname:  bytesToString(buf.Sysname[:]),
		Nodename: bytesToString(buf.Nodename[:]),
		Release:  bytesToString(buf.Release[:]),
		Version:  bytesToString(buf.Version[:]),
		Machine:  bytesToString(buf.Machine[:]),
	}, nil
}

func bytesToString(b []uint8) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	return string(b[:n])
}

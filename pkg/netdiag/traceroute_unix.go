//go:build darwin || freebsd || netbsd || openbsd
// +build darwin freebsd netbsd openbsd

package netdiag

import "syscall"

// Unix平台特定的常量定义
const (
	IPPROTO_ICMP = syscall.IPPROTO_ICMP
	SO_RCVTIMEO  = syscall.SO_RCVTIMEO
) 
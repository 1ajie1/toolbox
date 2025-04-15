//go:build linux
// +build linux

package netdiag

import "syscall"

// Linux平台特定的常量定义
const (
	IPPROTO_ICMP = syscall.IPPROTO_ICMP
	SO_RCVTIMEO  = syscall.SO_RCVTIMEO
) 
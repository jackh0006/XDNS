//go:build windows

// ==============================================================================
// XDNS
// Author: tajirax
// Github: https://github.com/WhiteDNS/XDNS
// Year: 2026
// ==============================================================================
package client

import "syscall"

func isConnBrokenErrno(errno syscall.Errno) bool {
	switch errno {
	case syscall.WSAECONNABORTED, syscall.WSAECONNRESET:
		return true
	}
	return false
}

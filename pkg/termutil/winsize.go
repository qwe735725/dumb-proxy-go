package termutil

import (
	"syscall"
	"unsafe"
)

// Winsize maps directly to the kernel's terminal geometry window structure
type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// GetSize queries the OS kernel for terminal columns and rows natively
func GetSize() (int, int) {
	ws := &Winsize{}
	
	// TIOCGWINSZ is the ioctl constant to Get Window Size on standard output (Fd 1)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	
	// In Go, if syscall returns an error code greater than 0, the operation failed
	if errno != 0 {
		return 80, 24 // Safe default fallback if terminal query fails
	}
	
	return int(ws.Col), int(ws.Row)
}


package main

import "syscall"

// rawRead issues read(2) directly on the descriptor, bypassing the poller the
// os package would otherwise install. That is what makes the goroutine
// genuinely uninterruptible rather than merely slow.
func rawRead(fd int, buf []byte) error {
	for {
		_, err := syscall.Read(fd, buf)
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}

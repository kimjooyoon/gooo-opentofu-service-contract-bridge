//go:build linux

package bridge

import "syscall"

func peakRSSKiB() (int, string) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, "UNKNOWN"
	}
	return int(usage.Maxrss), "MEASURED"
}

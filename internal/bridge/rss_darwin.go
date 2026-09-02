//go:build darwin

package bridge

import "syscall"

func peakRSSKiB() (int, string) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, "UNKNOWN"
	}
	return int(usage.Maxrss / 1024), "MEASURED"
}

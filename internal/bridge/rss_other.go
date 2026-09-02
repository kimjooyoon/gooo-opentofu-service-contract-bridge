//go:build !linux && !darwin

package bridge

func peakRSSKiB() (int, string) {
	return 0, "UNKNOWN"
}

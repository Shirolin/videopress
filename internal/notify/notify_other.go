//go:build !windows

package notify

func notify(title, message string) error { return nil }

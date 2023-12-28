//go:build !windows
// +build !windows

package mystring

func UTF8(raw string) string {
	return raw
}

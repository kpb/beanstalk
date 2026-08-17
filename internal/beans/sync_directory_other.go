//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package beans

func syncDirectory(string) error {
	return nil
}

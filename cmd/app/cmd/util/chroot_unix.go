package util

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

// Chroot chroot()s to the new path.
// NB: All file paths after this call are effectively relative to
// `rootfs`
func Chroot(rootfs string) error {
	if err := syscall.Chroot(rootfs); err != nil {
		return errors.Wrapf(err, "unable to chroot to %s", rootfs)
	}
	root := filepath.FromSlash("/")
	if err := os.Chdir(root); err != nil {
		return errors.Wrapf(err, "unable to chdir to %s", root)
	}
	return nil
}

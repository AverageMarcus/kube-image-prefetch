package copy

import (
	"io"
	"os"
)

// Run triggers the copy of this binary to the provided destination
func Run(dest string) error {
	original, err := os.Open(os.Args[0])
	if err != nil {
		return err
	}
	defer original.Close()

	new, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer new.Close()

	_, err = io.Copy(new, original)
	if err != nil {
		return err
	}

	return os.Chmod(dest, 0777)
}

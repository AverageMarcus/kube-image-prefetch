package copy

import (
	"io"
	"os"
)


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

	io.Copy(new, original)

	os.Chmod(dest, 0777)

	return nil
}

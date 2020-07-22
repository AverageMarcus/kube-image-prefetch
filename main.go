package main

import (
	"flag"
	"strings"

	"kube-image-prefetch/cmd/copy"
	"kube-image-prefetch/cmd/operator"
	"kube-image-prefetch/cmd/sleep"
)

func main() {
	command := flag.String("command", "operator", "The operation to perform [one of 'copy', 'sleep' or 'operator']")
	dest := flag.String("dest", "/mount/sleep", "The location to copy the binary to when command is 'copy'")
	flag.Parse()

	switch strings.ToLower(*command) {
	case "copy":
		if err := copy.Run(*dest); err != nil {
			panic(err)
		}
	case "sleep":
		if err := sleep.Run(); err != nil {
			panic(err)
		}
	default:
		if err := operator.Run(); err != nil {
			panic(err)
		}
	}
}

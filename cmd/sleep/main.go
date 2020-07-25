package sleep

import (
	"os"
	"os/signal"
	"syscall"
)

// Run triggers a sleep function that waits until a term/kill signal is received
func Run() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-c

	return nil
}

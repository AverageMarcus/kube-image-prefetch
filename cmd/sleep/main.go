package sleep

import (
	"os"
	"os/signal"
	"syscall"
)

func Run() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-c

	return nil
}

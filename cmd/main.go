package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baifei2014/go-mysql-ksync/conf"
	"github.com/baifei2014/go-mysql-ksync/library/log"
	"github.com/baifei2014/go-mysql-ksync/service"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Init(conf.Conf.Log)
	canal := service.NewCanal(conf.Conf)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-ch
		log.Info("ksync get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			canal.Close()
			log.Info("ksync exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

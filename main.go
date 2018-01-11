package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"flag"

	"github.com/oldtree/Axon/axonx"
	"github.com/oldtree/Axon/monitor"

	"github.com/oldtree/Axon/axonx/config"
	"github.com/oldtree/Axon/axonx/ins"
)

var instance = &ins.JsonProtocol{}

func Server() {
	s := axonx.NewAxonSrv(instance)
	s.Address = "127.0.0.1:8090"
	err := s.Start()
	if err != nil {
		fmt.Printf("error : %s", err.Error())
	}
	select {}
}

func Client() {
	var ext = make(chan struct{}, 1)
	cell, err := axonx.NewCellClientUseAddress(config.GetCfg().SrvAddress, instance, ext)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go cell.Active()
	time.Sleep(time.Second * 1)
	for index := 0; index < 10000000; index++ {
		time.Sleep(time.Second * 50)
		_, err := cell.Write([]byte(`{"msg":"golang"}`))
		if err != nil {
			break
		}
		//fmt.Printf("cell %s msg length : %d error : %v time %s \n", cell.LocalAddr().String(), n, err, time.Now().String())
	}
}

func JsonServer() {
	go Server()
	time.Sleep(time.Second * 5)
	for index := 0; index < 400; index++ {
		go Client()
		if index%50 == 0 {
			time.Sleep(time.Second * 5)
		}
	}
}

var conf = flag.String("c", "cfg.json", "config file")

func main() {
	flag.Parse()
	err := config.LoadCfg(*conf)
	if err != nil {
		os.Exit(-1)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	go axonx.DebugServer(config.GetCfg().DebugAddress)
	m := monitor.NewMonitor(config.GetCfg().Influx.Address, config.GetCfg().Influx.Username,
		config.GetCfg().Influx.Password, config.GetCfg().Influx.Databasename)
	go m.Dog()
	go JsonServer()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-sc
}

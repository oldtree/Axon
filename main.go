package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/oldtree/Axon/axonx"

	"github.com/oldtree/Axon/axonx/ins"
)

func Server() {
	s := axonx.NewAxonSrv(&ins.JsonProtocol{})
	s.Address = "127.0.0.1:8090"
	err := s.Start()
	if err != nil {
		fmt.Printf("error : %s", err.Error())
	}
	select {}
}

func Client() {
	var ext = make(chan int, 1)
	cell, err := axonx.NewCellClientUseAddress("127.0.0.1:8090", &ins.JsonProtocol{}, ext)
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
	for index := 0; index < 4000; index++ {
		go Client()
		if index%50 == 0 {
			time.Sleep(time.Second * 5)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	JsonServer()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-sc
}

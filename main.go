package main

import (
	"fmt"
	"log"
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
	cell.Write([]byte(`{"msg":"golang"}`))
	for index := 0; index < 10; index++ {
		cell.Write([]byte(`{"msg":"golang"}`))
	}
	select {}
}

func JsonServer() {
	go Server()
	time.Sleep(time.Second * 10)
	go Client()
	select {}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Println("Axonx")
	JsonServer()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-sc
}

package main

import (
	"flag"
	"fmt"
	"os"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/liusf/idgenerator/gen-go/idgenerator"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	port := flag.Int("p", 0, "port to listen to")
	help := flag.Bool("help", false, "show this help info")
	workerId := flag.Int("w", 0, "worker id (0-31)")
	datacenterId := flag.Int("dc", 0, "data center id (0-7)")
	flag.Parse()
	if *port <= 0 || *help {
		Usage()
		os.Exit(1)
	}

	protocolFactory := thrift.NewTCompactProtocolFactory()
	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	transport, err := thrift.NewTServerSocket(fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		fmt.Println("error open addr", err)
		return
	}

	handler, err := NewIdGeneratorHandler(int64(*workerId), int64(*datacenterId))
	if err != nil {
		fmt.Println("error starting server: ", err)
		os.Exit(1)
	}
	processor := idgenerator.NewIdGeneratorProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	err = server.Serve()
	if err != nil {
		fmt.Println("error running server: ", err)
	} else {
		fmt.Println("running id generator server")
	}
}

package main

import (
	"fmt"
	"net"
	"os"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/liusf/idgenerator/gen-go/idgenerator"
)

func newIdGeneratorClient(host string, port int) (timestamp int64, workerId int64) {
	trans, err := thrift.NewTSocket(net.JoinHostPort(host, fmt.Sprint(port)))
	if err != nil {
		fmt.Printf("Error resolving address %s:%d, %v\n", host, port, err)
		os.Exit(1)
	}
	framedTransport := thrift.NewTFramedTransport(trans)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	client := idgenerator.NewIdGeneratorClientFactory(framedTransport, protocolFactory)
	if err := trans.Open(); err != nil {
		fmt.Println("Error opening socket to 2", host, ":", port, " ", err)
		os.Exit(1)
	}
	timestamp, err1 := client.GetTimestamp()
	peerDatacenterId, err2 := client.GetDatacenterId()
	if err1 != nil || err2 != nil {
		fmt.Printf("Could not talk to peer %s:%d\n", host, port)
		os.Exit(1)
	}
	defer trans.Close()
	return timestamp, peerDatacenterId
}

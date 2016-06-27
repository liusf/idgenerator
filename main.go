package main

import (
	"flag"
	"fmt"
	"os"
  "math"
	"strings"
	"strconv"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/liusf/idgenerator/gen-go/idgenerator"
	"github.com/strava/go.serversets"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	port := flag.Int("p", 0, "port to listen to")
	help := flag.Bool("h", false, "show this help info")
	workerId := flag.Int("w", 0, "worker id (0-31)")
	datacenterId := flag.Int("dc", 0, "data center id (0-7)")
	zkServers := flag.String("zk", "", "check and register with zookeepers(ip:port,ip:port,..)")

	flag.Parse()
	if *port <= 0 || *help {
		Usage()
		os.Exit(1)
	}

	if *zkServers != "" {
		serversets.BaseDirectory = "/service"
		serversets.BaseZnodePath = func(environment serversets.Environment, service string) string {
			return serversets.BaseDirectory + "/" + service
		}
    addrs, serverSet := getPeerAddrs(*zkServers)
    sanityCheck(int64(*workerId), int64(*datacenterId), addrs)
    registerService(int64(*workerId), int(*port), serverSet)
		fmt.Println("Sanity check OK")
	}

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
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	processor := idgenerator.NewIdGeneratorProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	err = server.Serve()
	if err != nil {
		fmt.Println("error running server: ", err)
    os.Exit(1)
	} else {
		fmt.Println("running id generator server")
	}
}

func getPeerAddrs(zkServers string) ([]string, *serversets.ServerSet) {
	serverSet := serversets.New(serversets.Production, "idgenerators", strings.Split(zkServers, ","))
	watch, err := serverSet.Watch()
	if err != nil {
		fmt.Println("unable to connect to zk servers", zkServers, err)
		os.Exit(1)
	}
	endpoints := watch.Endpoints()
	fmt.Println("endpoints = ", endpoints)
	return endpoints, serverSet
}

func sanityCheck(workerId int64, datacenterId int64, addrs []string) {
	// check peers, not duplicated datacentId & workerId, not too much time shift
	if addrs == nil {
		fmt.Println("Unable to resolve peers address", addrs)
		os.Exit(1)
	}
	if len(addrs) == 0 {
		fmt.Println("No peers")
		return
	}
	var sumTimestamp int64 = 0
	for _, addr := range addrs {
		pair := strings.Split(addr, ":")
		port, err := strconv.Atoi(pair[1])
		if err != nil {
			fmt.Println("port error")
			continue
		}
		timestamp, peerDatacenterId, peerWorkerId := newIdGeneratorClient(pair[0], port)
		if datacenterId != peerDatacenterId {
			fmt.Printf("Worker at %s has datacenter_id %d, but ours is %d", addr, peerDatacenterId, datacenterId)
			os.Exit(1)
		} else if workerId == peerWorkerId {
      fmt.Println("Duplicated workerId", workerId)
      os.Exit(1)
    } else {
			sumTimestamp += timestamp
		}
	}
	avg := sumTimestamp / int64(len(addrs))
	if math.Abs(float64(avg-getTimestamp())) > 10000.0 {
		fmt.Printf("Timestamp sanity check failed. Mean timestamp is %d, but mine is %d, "+
			"so I'm more than 10s away from the mean", avg, getTimestamp())
		os.Exit(1)
	}
}

func registerService(workerId int64, port int, serverSet *serversets.ServerSet) {
  host := getHostname()
	_, err := serverSet.RegisterEndpoint(host, port, nil)
	if err != nil {
		fmt.Println("cannot register endpoint", err)
		os.Exit(1)
	}
}

func getHostname() string {
  host, err := os.Hostname()
  if err != nil {
    fmt.Println("cannot get local IPs")
    os.Exit(1)
  }
  return host
}
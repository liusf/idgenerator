package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"

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
	help := flag.Bool("h", false, "show this help info")
	workerId := flag.Int("w", 0, "worker id (0-31)")
	datacenterId := flag.Int("dc", 0, "data center id (0-7)")
	consulServers := flag.String("consul", "", "check peers with consul server hosts(ip:port,ip:port,...)")

	flag.Parse()
	if *port <= 0 || *help {
		Usage()
		os.Exit(1)
	}

	if *consulServers != "" {
		sanityCheck(int64(*workerId), int64(*datacenterId), *consulServers)
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
	} else {
		fmt.Println("running id generator server")
	}
}

type ServiceAddr struct {
	ServiceAddress string
	ServicePort    int
}

func sanityCheck(workerId int64, datacenterId int64, consulServers string) {
	// check peers, not duplicated datacentId & workerId, not too much time shift
	addrs := getPeerAddrs(consulServers)
	if addrs == nil {
		fmt.Println("Unable to resolve peers address", consulServers)
		os.Exit(1)
	}
	if len(addrs) == 0 {
		fmt.Println("No peers")
		return
	}
	var sumTimestamp int64
	for _, addr := range addrs {
		timestamp, peerDatacenterId := newIdGeneratorClient(addr.ServiceAddress, addr.ServicePort)
		if datacenterId != peerDatacenterId {
			fmt.Printf("Worker at %s has datacenter_id %d, but ours is %d", addr, peerDatacenterId, datacenterId)
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

func getPeerAddrs(consulServers string) []ServiceAddr {
	servers := strings.Split(consulServers, ",")
	var services []ServiceAddr
	for _, server := range servers {
		serviceUrl := fmt.Sprintf("http://%s/v1/catalog/service/idgenerator", server)
		resp, err := http.Get(serviceUrl)
		if err != nil {
			fmt.Println("get peers address error ", serviceUrl, err)
			continue
		} else {
			bytes, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Println("get service content error ", serviceUrl, err)
				continue
			} else {
				if err := json.Unmarshal(bytes, &services); err != nil {
					fmt.Println("service definition parse error", err)
					continue
				} else {
					return services
				}
			}
		}
	}
	return nil
}

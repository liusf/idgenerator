package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
  "math"
	"strings"
  "time"

  "git.apache.org/thrift.git/lib/go/thrift"
	"github.com/liusf/idgenerator/gen-go/idgenerator"
	"github.com/samuel/go-zookeeper/zk"
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
    addrs, zkConn := getPeerAddrs(*zkServers)
    sanityCheck(int64(*workerId), int64(*datacenterId), addrs)
    registerService(int64(*workerId), int(*port), zkConn)
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

func getPeerAddrs(zkServers string) ([]ServiceAddr, zk.Conn) {
  c, _, err := zk.Connect(strings.Split(zkServers, ","), time.Second)
  if err != nil {
    fmt.Println("unable to connect to zk servers ", zkServers)
    os.Exit(1)
  }
  children, _, err := c.Children("/service/idgenerators")
  if err != nil {
    fmt.Println("unable to get children ", zkServers)
    os.Exit(1)
  }
  var services []ServiceAddr = make([]ServiceAddr, 0)
  for _, node := range children {
    content, _, err := c.Get("/service/idgenerators/" + node)
    if err != nil {
      fmt.Println("unable to connect to node ", node)
      os.Exit(1)
    }
    var f ZkNode
    err = json.Unmarshal(content, &f)
    if err != nil {
      fmt.Println("unmarshal node json failed", err)
      os.Exit(1)
    }
    addr := ServiceAddr{f.ServiceEndpoint.Host, f.ServiceEndpoint.Port}
    services = append(services, addr)
  }
  return services, *c
}

type ServiceEndpoint struct {
  Host string `json:"host"`
  Port int    `json:"port"`
}

type ZkNode struct {
  ServiceEndpoint ServiceEndpoint `json:"serviceEndpoint"`
  AdditionalEndpoints interface{} `json:"additionalEndpoints,omitempty"`
  Status string                   `json:"status"`
  Shard int                       `json:"shard"`
}

type ServiceAddr struct {
	ServiceAddress string
	ServicePort    int
}

func sanityCheck(workerId int64, datacenterId int64, addrs []ServiceAddr) {
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
		timestamp, peerDatacenterId, peerWorkerId := newIdGeneratorClient(addr.ServiceAddress, addr.ServicePort)
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

func registerService(workerId int64, port int, zkConn zk.Conn) {
  host := getHostname()
  var s struct{}
  var endpoint = ZkNode{ServiceEndpoint:ServiceEndpoint{Host:host, Port:port}, AdditionalEndpoints:s, Status:"ALIVE", Shard:0}
  bytes, err := json.Marshal(endpoint)
  if err != nil {
    fmt.Println("json marshal failed")
    os.Exit(1)
  }
  path := fmt.Sprintf("/service/idgenerators/member_%d", workerId)
  _, err = zkConn.Create(path, bytes, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
  if err != nil {
    fmt.Println("reigster with zookeeper failed", err)
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
package main

import (
	"fmt"
	"sync"
	"time"
)

type IdGeneratorHandler struct {
	workerId     int64
	datacenterId int64
	generators   map[string]*IdGenerator
	mux          sync.Mutex
}

type IdGeneratorException struct {
	message string
}

func (exp IdGeneratorException) Error() string {
	return exp.message
}

type IdGenerator struct {
	workerId      int64
	datacenterId  int64
	scope         string
	sequenceId    int64
	lastTimestamp int64
	mux           sync.Mutex
}

// from snowflake
var epoch int64 = 1448899200000
var datacenterIdBits uint = 3
var workerIdBits uint = 4
var sequenceBits uint = 10

var maxDatacenterId int64 = -1 ^ (-1 << datacenterIdBits)
var maxWorkerId int64 = -1 ^ (-1 << workerIdBits)
var sequenceMask int64 = -1 ^ (-1 << sequenceBits)

var workerIdShift uint = sequenceBits
var datacenterIdShift uint = sequenceBits + workerIdBits
var timestampLeftShift uint = sequenceBits + workerIdBits + datacenterIdBits

func NewIdGeneratorHandler(workerId int64, datacenterId int64) (handler *IdGeneratorHandler, err error) {
	if workerId > maxWorkerId || workerId < 0 {
		err := newException(fmt.Sprintf("wrong worker id (must be in 0-%d)", maxWorkerId))
		return nil, err
	}
	if datacenterId > maxDatacenterId || datacenterId < 0 {
		err := newException(fmt.Sprintf("wrong data center id (must be in 0-%d)", maxDatacenterId))
		return nil, err
	}
	return &IdGeneratorHandler{workerId: workerId, datacenterId: datacenterId, generators: make(map[string]*IdGenerator)}, nil
}

func newException(message string) *IdGeneratorException {
	return &IdGeneratorException{message}
}

func newIdGenerator(workerId int64, datacenterId int64, scope string) *IdGenerator {
	return &IdGenerator{workerId: workerId, datacenterId: datacenterId, scope: scope, sequenceId: 0, lastTimestamp: -1}
}

func (p *IdGeneratorHandler) GetWorkerId() (r int64, err error) {
	return p.workerId, nil
}

func (p *IdGeneratorHandler) GetTimestamp() (r int64, err error) {
	return getTimestamp(), nil
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *IdGeneratorHandler) GetId(scope string) (r int64, err error) {
	p.mux.Lock()
	if x, found := p.generators[scope]; found {
		p.mux.Unlock()
		return x.nextId()
	} else {
		generator := newIdGenerator(p.workerId, p.datacenterId, scope)
		p.generators[scope] = generator
		p.mux.Unlock()
		return generator.nextId()
	}
}

func (p *IdGenerator) nextId() (r int64, err error) {
	p.mux.Lock()
	timestamp := getTimestamp()
	if timestamp < p.lastTimestamp {
		fmt.Printf("clock is moving backwards.  Rejecting requests until %d.", p.lastTimestamp)
		errMsg := fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", p.lastTimestamp-timestamp)
		err := newException(errMsg)
		defer p.mux.Unlock()
		return 0, err
	} else if timestamp == p.lastTimestamp {
		p.sequenceId = (p.sequenceId + 1) & sequenceMask
		if p.sequenceId == 0 {
			timestamp = tilNextMillis(p.lastTimestamp)
		}
	} else {
		p.sequenceId = 0
	}

	p.lastTimestamp = timestamp

	id := ((timestamp - epoch) << timestampLeftShift) |
		(p.datacenterId << datacenterIdShift) |
		(p.workerId << workerIdShift) |
		p.sequenceId
	defer p.mux.Unlock()
	return id, nil
}

func tilNextMillis(lastTimestamp int64) int64 {
	var timestamp = getTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = getTimestamp()
	}
	return timestamp
}

func (p *IdGeneratorHandler) GetDatacenterId() (r int64, err error) {
	return p.datacenterId, nil
}

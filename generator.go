package main

import (
	"fmt"
	"sync"
	"time"
)

type IdGeneratorException struct {
	message string
}

func (exp IdGeneratorException) Error() string {
	return exp.message
}

func newException(message string) *IdGeneratorException {
	return &IdGeneratorException{message}
}

type IdGenerator struct {
	workerId      int64
	datacenterId  int64
	scope         string
	sequenceId    int64
	lastTimestamp int64
	mux           sync.Mutex
}

func newIdGenerator(workerId int64, datacenterId int64, scope string) *IdGenerator {
	return &IdGenerator{workerId: workerId, datacenterId: datacenterId, scope: scope, sequenceId: 0, lastTimestamp: -1}
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

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

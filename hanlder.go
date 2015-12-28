package main

import (
	"fmt"
	"time"
)

type IdGeneratorHanlder struct {
	workerId     int64
	datacenterId int64
	sequenceId   int64
	scopes       map[string]int64
}

type IdGeneratorException struct {
	message string
}

func (exp IdGeneratorException) Error() string {
	return exp.message
}

// from snowflake
var epoch int64 = -2848899200000 //1448899200000
var datacenterIdBits uint = 3
var workerIdBits uint = 4
var sequenceBits uint = 10

var maxDatacenterId int64 = -1 ^ (-1 << datacenterIdBits)
var maxWorkerId int64 = -1 ^ (-1 << workerIdBits)
var sequenceMask int64 = -1 ^ (-1 << sequenceBits)

var workerIdShift uint = sequenceBits
var datacenterIdShift uint = sequenceBits + workerIdBits
var timestampLeftShift uint = sequenceBits + workerIdBits + datacenterIdBits

var lastTimestamp int64 = -1

func NewIdGeneratorHandler(workerId int64, datacentId int64) (handler *IdGeneratorHanlder, err error) {
	if workerId > maxWorkerId || workerId < 0 {
		err := newException(fmt.Sprintf("wrong worker id (must be in 0-%d)", maxWorkerId))
		return nil, err
	}
	if datacentId > maxDatacenterId || datacentId < 0 {
		err := newException(fmt.Sprintf("wrong data center id (must be in 0-%d)", maxDatacenterId))
		return nil, err
	}
	return &IdGeneratorHanlder{workerId, datacentId, 0, make(map[string]int64, 1)}, nil
}

func newException(message string) *IdGeneratorException {
	return &IdGeneratorException{message}
}

func (p *IdGeneratorHanlder) GetWorkerId() (r int64, err error) {
	return p.workerId, nil
}

func (p *IdGeneratorHanlder) GetTimestamp() (r int64, err error) {
	return getTimestamp(), nil
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *IdGeneratorHanlder) GetId(scope string) (r int64, err error) {

	timestamp := getTimestamp()
	if timestamp < lastTimestamp {
		fmt.Printf("clock is moving backwards.  Rejecting requests until %d.", lastTimestamp)
		errMsg := fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", lastTimestamp-timestamp)
		err := newException(errMsg)
		return 0, err
	} else if timestamp == lastTimestamp {
		p.sequenceId = (p.sequenceId + 1) & sequenceMask
		if p.sequenceId == 0 {
			timestamp = tilNextMillis(lastTimestamp)
		}
	} else {
		p.sequenceId = 0
	}

	lastTimestamp = timestamp

	id := ((timestamp - epoch) << timestampLeftShift) |
		(p.datacenterId << datacenterIdShift) |
		(p.workerId << workerIdShift) |
		p.sequenceId
	return id, nil
}

func tilNextMillis(lastTimestamp int64) int64 {
	var timestamp = getTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = getTimestamp()
	}
	return timestamp
}
func (p *IdGeneratorHanlder) GetDatacenterId() (r int64, err error) {
	return p.datacenterId, nil
}

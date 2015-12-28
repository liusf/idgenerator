package main

import (
  "time"
)

type IdGeneratorHanlder struct {
  workerId int64
  datacenterId int64
  scopes []string
}

func NewIdGeneratorHandler(workerId int64, datacentId int64) *IdGeneratorHanlder  {
  return &IdGeneratorHanlder{workerId, datacentId, make([]string, 1)}
}

func (p *IdGeneratorHanlder) GetWorkerId() (r int64, err error)  {
  return p.workerId, nil
}

func (p *IdGeneratorHanlder) GetTimestamp() (r int64, err error) {
  return makeTimestamp(), nil
}

func makeTimestamp() int64 {
  return time.Now().UnixNano() / int64(time.Millisecond)
}

func (p *IdGeneratorHanlder) GetId(scope string) (r int64, err error)  {
  return int64(111), nil
}

func (p *IdGeneratorHanlder) GetDatacenterId() (r int64, err error) {
  return int64(1), nil
}






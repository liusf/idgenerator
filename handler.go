package main

import (
	"fmt"
	"sync"
)

type IdGeneratorHandler struct {
	workerId     int64
	datacenterId int64
	generators   map[string]*IdGenerator
	mux          sync.Mutex
}

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

func (p *IdGeneratorHandler) GetWorkerId() (r int64, err error) {
	return p.workerId, nil
}

func (p *IdGeneratorHandler) GetTimestamp() (r int64, err error) {
	return getTimestamp(), nil
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

func (p *IdGeneratorHandler) GetDatacenterId() (r int64, err error) {
	return p.datacenterId, nil
}

func (p *IdGeneratorHandler) GetScopes() (r []string, err error)  {
	p.mux.Lock()
	keys := make([]string, len(p.generators))
	for d := range p.generators {
		keys = append(keys, d)
	}
	defer p.mux.Unlock()
	return keys, nil
}
package server

import (
	"config-consul/consul"
	"config-consul/utils"
	"github.com/cloudwego/kitex/server"
)

const (
	limiterConfigName = "limit"
)

// ConsulServerSuite etcd server config suite, configure limiter config dynamically from consul.
type ConsulServerSuite struct {
	uid          int64
	consulClient consul.Client
	service      string
	opts         utils.Options
}

// NewSuite service is the destination service.
func NewSuite(service string, cli consul.Client,
	opts ...utils.Option,
) *ConsulServerSuite {
	uid := consul.AllocateUniqueID()
	su := &ConsulServerSuite{
		uid:          uid,
		service:      service,
		consulClient: cli,
	}
	for _, opt := range opts {
		opt.Apply(&su.opts)
	}
	return su
}

// Options return a list client.Option
func (s *ConsulServerSuite) Options() []server.Option {
	opts := make([]server.Option, 0, 2)
	opts = append(opts, WithLimiter(s.service, s.consulClient, s.uid, s.opts))
	return opts
}

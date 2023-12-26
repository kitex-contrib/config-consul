package client

import (
	"config-consul/consul"
	"config-consul/utils"
	"github.com/cloudwego/kitex/client"
)

const (
	retryConfigName          = "retry"
	rpcTimeoutConfigName     = "rpc_timeout"
	circuitBreakerConfigName = "circuit_break"
)

type ConsulClientSuite struct {
	uid          int64
	consulClient consul.Client
	service      string
	client       string
	lconfig      *consul.ListenConfig
	opts         utils.Options
}

// NewSuite service is the destination service name and client is the local identity.
func NewSuite(service, client string, cli consul.Client,
	opts ...utils.Option,
) *ConsulClientSuite {
	uid := consul.AllocateUniqueID()
	su := &ConsulClientSuite{
		uid:          uid,
		service:      service,
		client:       client,
		consulClient: cli,
	}
	for _, opt := range opts {
		opt.Apply(&su.opts)
	}

	return su
}

// Options return a list client.Option
func (s *ConsulClientSuite) Options() []client.Option {
	opts := make([]client.Option, 0, 7)
	opts = append(opts, WithRetryPolicy(s.service, s.client, s.consulClient, s.uid, s.opts)...)
	opts = append(opts, WithRPCTimeout(s.service, s.client, s.consulClient, s.uid, s.opts)...)
	opts = append(opts, WithCircuitBreaker(s.service, s.client, s.consulClient, s.uid, s.opts)...)
	return opts
}

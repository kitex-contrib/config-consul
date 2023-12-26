package utils

import "config-consul/consul"

// Option is used to custom Options.
type Option interface {
	Apply(*Options)
}

// Options is used to initialize the nacos config suit or option.
type Options struct {
	ConsulCustomFunctions []consul.CustomFunction
}

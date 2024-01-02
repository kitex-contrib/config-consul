package server

import (
	"config-consul/consul"
	"config-consul/utils"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
	"sync/atomic"
)

// WithLimiter sets the limiter config from consul configuration center.
func WithLimiter(dest string, consulClient consul.Client, uniqueID int64, opts utils.Options) server.Option {
	param, err := consulClient.ServerConfigParam(&consul.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.ConsulCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	server.RegisterShutdownHook(func() {
		consulClient.DeregisterConfig(key, uniqueID)
	})
	return server.WithLimit(initLimitOptions(param.Type, key, uniqueID, consulClient))
}

func initLimitOptions(kind consul.ConfigType, key string, uniqueID int64, consulClient consul.Client) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[consul] %s server consul limiter updater init, config %v", key, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(data string, parser consul.ConfigParser) {
		lc := &limiter.LimiterConfig{}

		err := parser.Decode(kind, data, lc)
		if err != nil {
			klog.Warnf("[consul] %s server consul limiter config: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}

		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[consul] %s server consul limiter config failed as the updater is empty", key)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[consul] %s server consul limiter config: data %s may do not take affect", key, data)
		}
	}
	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)
	return opt
}

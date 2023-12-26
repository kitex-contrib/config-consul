package client

import (
	"config-consul/consul"
	"config-consul/utils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
)

// WithRPCTimeout sets the RPC timeout policy from etcd configuration center.
func WithRPCTimeout(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&consul.ConfigParamConfig{
		Category:          rpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}

	for _, f := range opts.ConsulCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(param.Type, key, dest, consulClient, uniqueID)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
	}
}

func initRPCTimeoutContainer(kind consul.ConfigType, key, dest string,
	consulClient consul.Client, uniqueID int64,
) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func(data string, parser consul.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		err := parser.Decode(kind, data, &configs)
		if err != nil {
			klog.Warnf("[consul] %s client consul rpc timeout: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}

		rpcTimeoutContainer.NotifyPolicyChange(configs)
	}

	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)

	return rpcTimeoutContainer
}

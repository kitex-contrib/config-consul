package client

import (
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/config-consul/consul"
	"github.com/kitex-contrib/config-consul/pkg/degradation"
	"github.com/kitex-contrib/config-consul/utils"
)

func WithDegradation(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&consul.ConfigParamConfig{
		Category:          degradationConfigName,
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
	container := initDegradationOptions(param.Type, key, dest, uniqueID, consulClient)
	return []client.Option{
		client.WithACLRules(container.GetAclRule()),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
	}
}

func initDegradationOptions(configType consul.ConfigType, key, dest string, uniqueID int64, consulClient consul.Client) *degradation.DegradationContainer {
	container := degradation.NewDegradationContainer()
	onChangeCallback := func(data string, parser consul.ConfigParser) {
		config := &degradation.DegradationConfig{}
		err := parser.Decode(configType, data, config)
		if err != nil {
			klog.Warnf("[consul] %s server consul degradation config: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}
		container.NotifyPolicyChange(config)
	}
	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)
	return container
}

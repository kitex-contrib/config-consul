package client

import (
	"config-consul/consul"
	"config-consul/utils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/retry"
)

// WithRetryPolicy sets the retry policy from etcd configuration center.
func WithRetryPolicy(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&consul.ConfigParamConfig{
		Category:          retryConfigName,
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
	rc := initRetryContainer(param.Type, key, dest, consulClient, uniqueID)
	return []client.Option{
		client.WithRetryContainer(rc),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
		client.WithCloseCallbacks(rc.Close),
	}
}

func initRetryContainer(kind consul.ConfigType, key, dest string,
	consulClient consul.Client, uniqueID int64,
) *retry.Container {
	retryContainer := retry.NewRetryContainerWithPercentageLimit()

	ts := utils.ThreadSafeSet{}

	onChangeCallback := func(data string, parser consul.ConfigParser) {
		// the key is method name, wildcard "*" can match anything.
		rcs := map[string]*retry.Policy{}
		err := parser.Decode(kind, data, &rcs)
		if err != nil {
			klog.Warnf("[consul] %s client etcd retry: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}
		set := utils.Set{}
		for method, policy := range rcs {
			set[method] = true
			if policy.Enable && policy.BackupPolicy == nil && policy.FailurePolicy == nil {
				klog.Warnf("[consul] %s client policy for method %s BackupPolicy and FailurePolicy must not be empty at same time",
					dest, method)
				continue
			}
			retryContainer.NotifyPolicyChange(method, *policy)
		}

		for _, method := range ts.DiffAndEmplace(set) {
			retryContainer.DeletePolicy(method)
		}
	}

	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)

	return retryContainer
}

// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consul

import (
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.uber.org/zap"

	cwConsul "github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	cwUtils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"
)

const WatchByKey = "key"

type Key struct {
	Type   ConfigType
	Prefix string
	Path   string
}

type ListenConfig struct {
	Key        string
	Type       string
	DataCenter string
	Token      string
	ConsulAddr string
	Namespace  string
	Partition  string
}
type Client interface {
	SetParser(configParser ConfigParser)
	ClientConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	ServerConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	RegisterConfigCallback(key string, uniqueID int64, callback func(string, ConfigParser))
	DeregisterConfig(key string, uniqueID int64)
}

type Options struct {
	Addr             string
	Prefix           string
	ServerPathFormat string
	ClientPathFormat string
	DataCenter       string
	TimeOut          time.Duration
	NamespaceId      string
	Token            string
	Partition        string
	LoggerConfig     *zap.Config
	ConfigParser     ConfigParser
}

type client struct {
	cwClient cwConsul.Client
}

func NewClient(opts Options) (Client, error) {
	c, err := cwConsul.NewClient(*transferOpinion(opts))
	if err != nil {
		return nil, err
	}

	return &client{cwClient: c}, nil
}

// SetParser support customise parser
func (c *client) SetParser(parser ConfigParser) {
	cwParser, err := transferConfigParser(parser)
	if err != nil {
		klog.Errorf("SetParser failed,error: %s", err.Error())
		return
	}

	c.cwClient.SetParser(cwParser)
}

func transferCfs(cfs ...CustomFunction) []cwConsul.CustomFunction {
	cwCfs := make([]cwConsul.CustomFunction, 0, len(cfs))
	for _, cf := range cfs {
		cwCfs = append(cwCfs, func(key *cwConsul.Key) {
			cf(transferCwKey(key))
		})
	}
	return cwCfs
}

func (c *client) ClientConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	cwKey, err := c.cwClient.ClientConfigParam(&cpc.CwConfigParamConfig, transferCfs(cfs...)...)

	if err != nil {
		return Key{}, err
	}

	return *transferCwKey(&cwKey), nil
}

func (c *client) ServerConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	cwKey, err := c.cwClient.ServerConfigParam(&cpc.CwConfigParamConfig, transferCfs(cfs...)...)

	if err != nil {
		return Key{}, err
	}

	return *transferCwKey(&cwKey), nil
}

// RegisterConfigCallback register the callback function to consul client.
func (c *client) RegisterConfigCallback(key string, uniqueID int64, callback func(string, ConfigParser)) {
	cwCallback := func(key string, cwParser cwUtils.ConfigParser) {
		parser, err := transferCwConfigParser(cwParser)
		if err != nil {
			klog.Errorf("transferCwConfigParser failed,error: %s", err.Error())
			return
		}
		callback(key, parser)
	}
	c.cwClient.RegisterConfigCallback(key, uniqueID, cwCallback)
}

func (c *client) DeregisterConfig(key string, uniqueID int64) {
	c.cwClient.DeregisterConfig(key, uniqueID)
}

// func transferKey(key *Key) *cwConsul.Key {
// 	return &cwConsul.Key{
// 		Type:   cwUtils.ConfigType(key.Type),
// 		Prefix: key.Prefix,
// 		Path:   key.Path,
// 	}
// }

func transferCwKey(key *cwConsul.Key) *Key {
	return &Key{
		Type:   ConfigType(key.Type),
		Prefix: key.Prefix,
		Path:   key.Path,
	}
}

func transferOpinion(opinion Options) *cwConsul.Options {
	cwConfigParser, err := transferConfigParser(opinion.ConfigParser)
	if err != nil {
		klog.Errorf("transferConfigParser failed,error: %s", err.Error())
		return &cwConsul.Options{}
	}

	return &cwConsul.Options{
		Addr:             opinion.Addr,
		Prefix:           opinion.Prefix,
		ServerPathFormat: opinion.ServerPathFormat,
		ClientPathFormat: opinion.ClientPathFormat,
		DataCenter:       opinion.DataCenter,
		TimeOut:          opinion.TimeOut,
		NamespaceId:      opinion.NamespaceId,
		Token:            opinion.Token,
		Partition:        opinion.Partition,
		LoggerConfig:     opinion.LoggerConfig,
		ConfigParser:     cwConfigParser,
	}
}

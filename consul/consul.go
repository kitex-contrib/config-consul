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
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
)

const WatchByKey = consul.WatchByKey

type Key = consul.Key

type ListenConfig = consul.ListenConfig

type Client = consul.Client

type Options = consul.Options

var _ Client = &client{}

type client struct {
	cwClient consul.Client
}

func NewClient(opts Options) (Client, error) {
	return consul.NewClient(opts)
}

// SetParser support customise parser
func (c *client) SetParser(parser ConfigParser) {
	c.cwClient.SetParser(parser)
}

func (c *client) ClientConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.cwClient.ClientConfigParam(cpc, cfs...)
}

func (c *client) ServerConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.cwClient.ServerConfigParam(cpc, cfs...)
}

// RegisterConfigCallback register the callback function to consul client.
func (c *client) RegisterConfigCallback(key string, uniqueID int64, callback func(string, ConfigParser)) {
	c.cwClient.RegisterConfigCallback(key, uniqueID, callback)
}

func (c *client) DeregisterConfig(key string, uniqueID int64) {
	c.cwClient.DeregisterConfig(key, uniqueID)
}

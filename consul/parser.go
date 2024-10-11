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

	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/utils"
)

type ConfigType = utils.ConfigType

const (
	JSON                      ConfigType = "json"
	YAML                      ConfigType = "yaml"
	HCL                       ConfigType = "hcl"
	ConsulDefaultConfigAddr              = "127.0.0.1:8500"
	ConsulDefaultConfiGPrefix            = "KitexConfig"
	ConsulDefaultTimeout                 = 5 * time.Second
	ConsulDefaultDataCenter              = "dc1"
	ConsulDefaultClientPath              = "{{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}}"
	ConsulDefaultServerPath              = "{{.ServerServiceName}}/{{.Category}}"
)

var _ ConfigParser = &parser{}

// CustomFunction use for customize the config parameters.
type CustomFunction = consul.CustomFunction

type ConfigParamConfig = utils.ConfigParamConfig

type ConfigParser = utils.ConfigParser

type parser struct {
	cwParser utils.ConfigParser
}

func (p *parser) Decode(configType ConfigType, data string, config interface{}) error {
	return p.cwParser.Decode(utils.ConfigType(configType), data, config)
}

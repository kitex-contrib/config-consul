package consul

import (
	"encoding/json"
	"fmt"
)

type ConfigType string

const (
	JSON                         ConfigType = "json"
	YAML                         ConfigType = "yaml"
	HCL                          ConfigType = "hcl"
	ConsulDefaultConfigServerURL            = "127.0.0.1:8500"
	ConsulDefaultDateCenter                 = "dc1"
)

type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerviceName    string
}

type ConfigParser interface {
	Decode(data string, config interface{}) error
}

type parser struct{}

func (p *parser) Decode(kind ConfigType, data string, config interface{}) error {
	//hclParser, err := hcl.Parse(hclString)
	switch kind {
	case JSON, YAML:
		return json.Unmarshal([]byte(data), config)
	case HCL:
		//待补充
		return nil
	default:
		return fmt.Errorf("unsupported config data type %s", kind)
	}
}

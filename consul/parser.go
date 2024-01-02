package consul

import (
	"encoding/json"
	"fmt"
	"time"
)

type ConfigType string

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
type CustomFunction func(*Key)

type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}

type ConfigParser interface {
	Decode(kind ConfigType, data string, config interface{}) error
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

func defaultConfigParse() ConfigParser {
	return &parser{}
}

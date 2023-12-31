package consul

import (
	"bytes"
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
	"html/template"
	"strconv"
	"sync"
	"time"
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
	Datacenter string
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
	consulCli          *api.Client
	lconfig            *ListenConfig
	parser             ConfigParser
	consulTimeout      time.Duration
	prefixTemplate     *template.Template
	serverPathTemplate *template.Template
	clientPathTemplate *template.Template
	cancelMap          map[string]context.CancelFunc
	m                  sync.Mutex
}

func NewClient(opts Options) (Client, error) {
	if opts.Addr == "" {
		opts.Addr = ConsulDefaultConfigAddr
	}
	if opts.Prefix == "" {
		opts.Prefix = ConsulDefaultConfiGPrefix
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = defaultConfigParse()
	}
	if opts.TimeOut == 0 {
		opts.TimeOut = ConsulDefaultTimeout
	}
	if opts.ClientPathFormat == "" {
		opts.ClientPathFormat = ConsulDefaultClientPath
	}
	if opts.ServerPathFormat == "" {
		opts.ServerPathFormat = ConsulDefaultServerPath
	}
	if opts.DataCenter == "" {
		opts.DataCenter = ConsulDefaultDataCenter
	}
	consulClient, err := api.NewClient(&api.Config{
		Address:    opts.Addr,
		Datacenter: opts.DataCenter,
		Token:      opts.Token,
		Namespace:  opts.NamespaceId,
		Partition:  opts.Partition,
	})
	if err != nil {
		return nil, err
	}
	prefixTemplate, err := template.New("prefix").Parse(opts.Prefix)
	if err != nil {
		return nil, err
	}
	serverNameTemplate, err := template.New("serverName").Parse(opts.ServerPathFormat)
	if err != nil {
		return nil, err
	}
	clientNameTemplate, err := template.New("clientName").Parse(opts.ClientPathFormat)
	if err != nil {
		return nil, err
	}
	lconfig := &ListenConfig{
		Type:       WatchByKey,
		Datacenter: opts.DataCenter,
		Token:      opts.Token,
		ConsulAddr: opts.Addr,
		Namespace:  opts.NamespaceId,
		Partition:  opts.Partition,
	}
	c := &client{
		consulCli:          consulClient,
		parser:             opts.ConfigParser,
		consulTimeout:      opts.TimeOut,
		prefixTemplate:     prefixTemplate,
		serverPathTemplate: serverNameTemplate,
		clientPathTemplate: clientNameTemplate,
		lconfig:            lconfig,
		cancelMap:          make(map[string]context.CancelFunc),
	}
	return c, nil
}

// SetParser support customise parser
func (c *client) SetParser(parser ConfigParser) {
	c.parser = parser
}

func (c *client) ClientConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.configParam(cpc, c.clientPathTemplate, cfs...)
}

func (c *client) ServerConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.configParam(cpc, c.serverPathTemplate, cfs...)
}

// configParam render config parameters. All the parameters can be customized with CustomFunction.
// ConfigParam explain:
//  1. Prefix: KitexConfig by default.
//  2. ServerPath: {{.ServerServiceName}}/{{.Category}} by default.
//     ClientPath: {{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}} by default.
func (c *client) configParam(cpc *ConfigParamConfig, t *template.Template, cfs ...CustomFunction) (Key, error) {
	param := Key{Type: JSON}
	var err error
	param.Path, err = c.render(cpc, t)
	if err != nil {
		return param, err
	}
	param.Prefix, err = c.render(cpc, c.prefixTemplate)
	if err != nil {
		return param, err
	}

	for _, cf := range cfs {
		cf(&param)
	}
	return param, nil
}

func (c *client) render(cpc *ConfigParamConfig, t *template.Template) (string, error) {
	var tpl bytes.Buffer
	err := t.Execute(&tpl, cpc)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// RegisterConfigCallback register the callback function to consul client.
func (c *client) RegisterConfigCallback(key string, uniqueID int64, callback func(string, ConfigParser)) {
	go func() {
		clientCtx, cancel := context.WithCancel(context.Background())
		params := make(map[string]interface{})
		params["datacenter"] = c.lconfig.Datacenter
		params["token"] = c.lconfig.Token
		params["type"] = c.lconfig.Type
		c.lconfig.Key = key
		c.registerCancelFunc(key, uniqueID, cancel)
		w, err := watch.Parse(params)

		if err != nil {
			klog.Debugf("[consul] key:add listen for %s failed", key)
		}
		w.Handler = func(u uint64, i interface{}) {
			kv := i.(*api.KVPair)
			v := string(kv.Value)
			klog.Debugf("[consul] config key:%s listen for %s failed", key)
			callback(v, c.parser)
		}
		go func() {
			err := w.Run(c.lconfig.ConsulAddr)
			if err != nil {
				klog.Errorf("[consul] listen key: %s failed,error: %s", key, err.Error())
			}
		}()
		for {
			select {
			case <-clientCtx.Done():
				w.Stop()
				return
			default:

			}
		}
	}()
	_, cancel := context.WithTimeout(context.Background(), c.consulTimeout)
	defer cancel()
	kv := c.consulCli.KV()
	get, _, err := kv.Get(c.lconfig.Key, nil)
	if err != nil {
		klog.Debugf("[consul] key: %s config get value failed", get.Key)
		return
	}
	if len(get.Value) == 0 {
		return
	}
	callback(string(get.Value), c.parser)
}

func (c *client) DeregisterConfig(key string, uniqueID int64) {
	c.deregisterCancelFunc(key, uniqueID)
}

func (c *client) deregisterCancelFunc(key string, uniqueID int64) {
	c.m.Lock()
	clientKey := key + "/" + strconv.FormatInt(uniqueID, 10)
	cancel := c.cancelMap[clientKey]
	cancel()
	delete(c.cancelMap, clientKey)
	c.m.Unlock()
}

func (c *client) registerCancelFunc(key string, uniqueID int64, cancel context.CancelFunc) {
	c.m.Lock()
	clientKey := key + "/" + strconv.FormatInt(uniqueID, 10)
	c.cancelMap[clientKey] = cancel
	c.m.Unlock()
}

package goconfig_center_apollo

import (
	"github.com/nova2018/goconfig"
	gocenter "github.com/nova2018/goconfig-center"
	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/spf13/viper"
	"strings"
)

type apolloConfig struct {
	gocenter.ConfigDriver `mapstructure:",squash"`
	Prefix                string `mapstructure:"prefix"`
	AppId                 string `mapstructure:"app_id"`
	Type                  string `mapstructure:"type"`
	Endpoint              string `mapstructure:"endpoint"`
	Namespace             string `mapstructure:"namespace"`
	Secret                string `mapstructure:"secret"`
}

type apolloDriver struct {
	cfgViper *viper.Viper
	cfg      *apolloConfig
	v        []*viper.Viper
	goConfig *goconfig.Config
}

func (a *apolloDriver) Name() string {
	return a.cfg.Driver
}

func (a *apolloDriver) IsSame(viper *viper.Viper) bool {
	return goconfig.Equal(a.cfgViper, viper)
}

func (a *apolloDriver) Watch() bool {
	for _, x := range a.v {
		a.goConfig.AddWatchViper(goconfig.WatchRemote, x, a.Prefix())
	}
	return true
}

func (a *apolloDriver) Unwatch() bool {
	for _, x := range a.v {
		a.goConfig.DelViper(x)
	}
	return true
}

func (a *apolloDriver) Prefix() string {
	return a.cfg.Prefix
}

func Factory(config *goconfig.Config, cfg *viper.Viper) (gocenter.Driver, error) {
	var c apolloConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}

	remote.SetAppID(c.AppId)
	if c.Type == "" {
		c.Type = "prop"
	}
	listNamespace := strings.Split(c.Namespace, ",")
	listV := make([]*viper.Viper, 0, len(listNamespace))
	isHit := make(map[string]bool, len(listNamespace))
	for _, namespace := range listNamespace {
		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			continue
		}
		if isHit[namespace] {
			continue
		}
		isHit[namespace] = true
		v := viper.New()
		v.SetConfigType(c.Type)
		if c.Secret == "" {
			if err := v.AddRemoteProvider("apollo", c.Endpoint, namespace); err != nil {
				return nil, err
			}
		} else {
			if err := v.AddSecureRemoteProvider("apollo", c.Endpoint, namespace, c.Secret); err != nil {
				return nil, err
			}
		}
		listV = append(listV, v)
	}

	return &apolloDriver{
		cfgViper: cfg,
		cfg:      &c,
		v:        listV,
		goConfig: config,
	}, nil
}

func init() {
	gocenter.Register("apollo", Factory)
}

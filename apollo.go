package goconfig_center_apollo

import (
	"github.com/nova2018/goconfig"
	gocenter "github.com/nova2018/goconfig-center"
	"github.com/shima-park/agollo"
	remote "github.com/shima-park/agollo/viper-remote"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type apolloConfig struct {
	gocenter.ConfigDriver `mapstructure:",squash"`
	Prefix                string `mapstructure:"prefix"`
	AppId                 string `mapstructure:"appId"`
	Endpoint              string `mapstructure:"endpoint"`
	Namespace             string `mapstructure:"namespace"`
	Cluster               string `mapstructure:"cluster"`
	AccessKey             string `mapstructure:"accessKey"`
}

func getConfigType(namespace string) string {
	ext := filepath.Ext(namespace)

	if len(ext) > 1 {
		fileExt := ext[1:]
		// 还是要判断一下碰到，TEST.Namespace1
		// 会把Namespace1作为文件扩展名
		for _, e := range viper.SupportedExts {
			if e == fileExt {
				return fileExt
			}
		}
	}

	return "properties"
}

type apolloDriver struct {
	cfg      *apolloConfig
	v        []*viper.Viper
	goConfig *goconfig.Config
}

func (a *apolloDriver) Name() string {
	return a.cfg.Driver
}

func (a *apolloDriver) Watch() bool {
	for _, x := range a.v {
		a.goConfig.AddWatchViper(goconfig.WatchRemote, x, a.cfg.Prefix)
	}
	return true
}

func (a *apolloDriver) Unwatch() bool {
	for _, x := range a.v {
		a.goConfig.DelViper(x)
	}
	return true
}

func apolloFactory(config *goconfig.Config, cfg *viper.Viper) (gocenter.Driver, error) {
	var c apolloConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}

	remote.SetAppID(c.AppId)
	listNamespace := strings.Split(c.Namespace, ",")
	listV := make([]*viper.Viper, 0, len(listNamespace))
	isHit := make(map[string]bool, len(listNamespace))
	opts := []agollo.Option{
		agollo.AutoFetchOnCacheMiss(),
		agollo.FailTolerantOnBackupExists(),
	}
	if c.Cluster != "" {
		opts = append(opts, agollo.Cluster(c.Cluster))
	}
	if c.AccessKey != "" {
		opts = append(opts, agollo.WithClientOptions(agollo.WithAccessKey(c.AccessKey)))
	}
	remote.SetAgolloOptions(opts...)
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
		cType := getConfigType(namespace)
		v.SetConfigType(cType)
		if err := v.AddRemoteProvider("apollo", c.Endpoint, namespace); err != nil {
			return nil, err
		}
		listV = append(listV, v)
	}

	return &apolloDriver{
		cfg:      &c,
		v:        listV,
		goConfig: config,
	}, nil
}

func init() {
	gocenter.Register("apollo", apolloFactory)
}

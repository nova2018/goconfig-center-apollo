package goconfig_center_apollo

import (
	"bytes"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/nova2018/goconfig"
	gocenter "github.com/nova2018/goconfig-center"
	"github.com/shima-park/agollo"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type apolloConfig struct {
	gocenter.ConfigDriver `mapstructure:",squash"`
	Prefix                string `mapstructure:"prefix"`
	AppId                 string `mapstructure:"appId"`
	Endpoint              string `mapstructure:"endpoint"`
	Namespace             string `mapstructure:"namespace"`
	Cluster               string `mapstructure:"cluster"`
	SLB                   bool   `mapstructure:"slb"`
	AccessKey             string `mapstructure:"accessKey"`
	IP                    string `mapstructure:"ip"`
}

type apolloDriver struct {
	goconfig *goconfig.Config
	cfg      *apolloConfig
	client   agollo.Agollo
	onChange chan struct{}
	closed   bool
	v        *viper.Viper
	dirty    bool
	lock     sync.Mutex
	cLock    sync.RWMutex // lock for protected onChange channel
}

func (a *apolloDriver) GetViper() (*viper.Viper, error) {
	if !a.dirty && a.v != nil {
		return a.v, nil
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	listNamespace := a.client.Options().PreloadNamespaces
	v := viper.New()
	for _, namespace := range listNamespace {
		cfg := a.client.GetNameSpace(namespace)
		cType := getConfigType(namespace)
		v.SetConfigType(cType)
		if cType == "properties" {
			b, err := marshalProperties(cfg)
			if err != nil {
				return a.v, err
			}
			_ = v.ReadConfig(bytes.NewReader(b))
		} else {
			if content, ok := cfg["content"]; ok {
				v.SetConfigType(cType)
				_ = v.ReadConfig(bytes.NewBufferString(content.(string)))
			}
		}
	}
	a.v = v
	a.dirty = false
	return v, nil
}

func marshalProperties(c map[string]interface{}) ([]byte, error) {
	p := properties.NewProperties()
	for key, val := range c {
		_, _, err := p.Set(key, fmt.Sprint(val))
		if err != nil {
			return nil, err
		}
	}
	buff := bytes.NewBuffer(nil)
	_, err := p.WriteComment(buff, "#", properties.UTF8)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
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

func (a *apolloDriver) OnChange() <-chan struct{} {
	if a.onChange == nil && !a.closed {
		a.cLock.Lock()
		a.onChange = make(chan struct{})
		a.cLock.Unlock()
		go func() {
			errCh := a.client.Start()
			watchCh := a.client.Watch()
			go func(errCh <-chan *agollo.LongPollerError, watchCh <-chan *agollo.ApolloResponse) {
				for !a.closed {
					select {
					case err := <-errCh:
						fmt.Printf("%s apollo listen failure! err=%s\n", time.Now().Format("2006-01-02 15:04:05"), err.Err)
					case w := <-watchCh:
						if w.Changes.Len() > 0 {
							// 广播通知
							a.cLock.RLock()
							if a.onChange != nil {
								a.lock.Lock()
								a.dirty = true
								a.lock.Unlock()
								a.onChange <- struct{}{}
							}
							a.cLock.RUnlock()
						}
					}
				}
				a.client.Stop()
			}(errCh, watchCh)
		}()
	}
	return a.onChange
}

func (a *apolloDriver) Name() string {
	return a.cfg.Driver
}

func (a *apolloDriver) Watch() bool {
	if !a.closed {
		a.goconfig.AddCustomWatchViper(a, a.cfg.Prefix)
	}
	return true
}

func (a *apolloDriver) Unwatch() bool {
	if !a.closed {
		a.closed = true
		a.cLock.Lock()
		if a.onChange != nil {
			close(a.onChange)
			a.onChange = nil
		}
		a.cLock.Unlock()
		a.goconfig.DelViper(a)
	}
	return true
}

func Factory(config *goconfig.Config, cfg *viper.Viper) (gocenter.Driver, error) {
	var c apolloConfig
	if err := cfg.Unmarshal(&c); err != nil {
		return nil, err
	}

	listNamespace := strings.Split(c.Namespace, ",")
	listOpts := append([]agollo.Option{
		agollo.PreloadNamespaces(listNamespace...),
		agollo.FailTolerantOnBackupExists(),
	})
	if c.Cluster != "" {
		listOpts = append(listOpts, agollo.Cluster(c.Cluster))
	}
	if c.AccessKey != "" {
		listOpts = append(listOpts, agollo.AccessKey(c.AccessKey))
	}
	if c.SLB {
		listOpts = append(listOpts, agollo.EnableSLB(true))
	}
	if c.IP != "" {
		listOpts = append(listOpts, agollo.WithClientOptions(agollo.WithIP(c.IP)))
	}
	a, err := agollo.New(c.Endpoint, c.AppId, listOpts...)
	if err != nil {
		return nil, err
	}

	return &apolloDriver{
		goconfig: config,
		client:   a,
		cfg:      &c,
	}, nil
}

func init() {
	gocenter.Register("apollo", Factory)
}

package goconfig_center_apollo

import (
	"bytes"
	"fmt"
	gocenter "github.com/nova2018/goconfig-center"
	"github.com/spf13/viper"
	"testing"
	"time"
)

func TestApollo(t *testing.T) {
	toml := []byte(`
[config_center]
enable = true
[[config_center.drivers]]
enable = true
driver = "apollo"
endpoint = "http://localhost:8080/"
appId = "SampleApp"
namespace = "application"
`)
	v := viper.New()
	v.SetConfigType("toml")
	_ = v.ReadConfig(bytes.NewBuffer(toml))

	fmt.Println("v.AllSettings:", v.AllSettings())

	//gconfig := goconfig.New()
	//gconfig.AddNoWatchViper(v)
	//gg := gocenter.New(gconfig)
	//gg.Watch()
	//gconfig.StartWait()

	center := gocenter.NewWithViper(v)
	center.Watch()
	gconfig := center.GetConfig()

	gconfig.OnKeyChange("xxx", func() {
		fmt.Println("key update!!!")
	})

	for {
		time.Sleep(1 * time.Second)
		fmt.Printf("%v 1app.AllSettings:%v\n", time.Now(), gconfig.GetConfig().AllSettings())
	}
}

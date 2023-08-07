package goconfig_center_apollo

import (
	"bytes"
	"fmt"
	"github.com/nova2018/goconfig"
	gocenter "github.com/nova2018/goconfig-center"
	"github.com/spf13/viper"
	"regexp"
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
[[a.aa]]
x=1
[[a.aa]]
x=2
[b.c.d]
x=1
[b.c.3]
x=2
`)
	v := viper.New()
	v.SetConfigType("toml")
	_ = v.ReadConfig(bytes.NewBuffer(toml))

	fmt.Println("v.AllSettings:", v.AllSettings())
	fmt.Println("v.AllKeys:", v.AllKeys())

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

	gconfig.OnMapKeyChange("abc", func(e goconfig.ConfigUpdateEvent) {
		fmt.Println(e)
	})

	mKey, _ := regexp.Compile(`^x\.1\.abc`)
	gconfig.OnMatchKeyChange(mKey, func(e goconfig.ConfigUpdateEvent) {
		fmt.Println(e)
	})

	for {
		time.Sleep(1 * time.Second)
		fmt.Printf("%v 1app.AllSettings:%v\n", time.Now(), gconfig.GetConfig().AllSettings())
	}
}

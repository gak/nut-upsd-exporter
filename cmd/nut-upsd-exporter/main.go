package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"nut-upsd-exporter"
	"time"
)

type Config struct {
	Bind       string `env:"NUE_BIND" default:":8080"`
	UPSDHostIP string `env:"NUE_UPSD_HOSTIP"`
	UPSDDevice string `env:"NUE_UPSD_DEVICE"`
}

func main() {
	config := Config{}
	kong.Parse(&config)

	e := nut.Exporter{
		Bind: config.Bind,
		UPSC: nut.UPSClient{
			HostIP: config.UPSDHostIP,
			Name:   config.UPSDDevice,
		},
	}

	if err := e.Init(); err != nil {
		panic(err)
	}

	go func() {
		for {
			fmt.Println("polling")
			if err := e.Poll(); err != nil {
				panic(err)
			}

			time.Sleep(10 * time.Second)
		}
	}()

	if err := e.Listen(); err != nil {
		panic(err)
	}

	select {}
}

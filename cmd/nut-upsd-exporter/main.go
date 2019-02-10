package main

import (
	"fmt"
	"nut-upsd-exporter"
	"time"
)

func main() {
	e := nut.Exporter{
		Bind: ":31515",
		UPSC: nut.UPSClient{
			HostIP: "10.0.0.3:3493",
			Name:   "eaton",
		},
	}

	if err := e.Init(); err != nil {
		panic(err)
	}

	go func() {
		for {
			fmt.Println("Polling")
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

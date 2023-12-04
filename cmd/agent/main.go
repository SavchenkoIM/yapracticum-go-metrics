package main

import (
	"metricsPoll"
	"time"
)

func main() {
	mh := metricsPoll.NewMetricsHandler()
	ctr := 0

	for {
		mh.RefreshData()
		time.Sleep(2 * time.Second)
		ctr += 1
		if ctr >= 5 {
			mh.SendData()
			ctr = 0
		}
	}
}

package internal

import (
	"fmt"
	"os"
	"time"
)

var NETWORKS = []string{"qa", "testing", "production"}

var SUBSTRATE_URL = map[string][]string{
	"qa":         {"wss://tfchain.qa.grid.tf/ws"},
	"testing":    {"wss://tfchain.test.grid.tf/ws"},
	"production": {"wss://tfchain.grid.tf/ws"},
}

type worker struct {
	src string
	dst string

	// optional
	interval     time.Duration
	substrateUrl map[string][]string

	substrateClient substrateClient
}

// new instance of the worker
func NewWorker(src string, dst string, params map[string]any) worker {
	interval := time.Minute * 10
	substrateUrl := SUBSTRATE_URL

	if val, ok := params["interval"]; ok {
		interval = val.(time.Duration)
	}

	if val, ok := params["qa"]; ok {
		substrateUrl["qa"] = val.([]string)
	}

	if val, ok := params["testing"]; ok {
		substrateUrl["testing"] = val.([]string)
	}

	if val, ok := params["production"]; ok {
		substrateUrl["production"] = val.([]string)
	}

	return worker{
		src:          src,
		dst:          dst,
		interval:     interval,
		substrateUrl: substrateUrl,
	}
}

// updateZosVersion updates the latest zos flist for a specific network with the updated zos version
func (w *worker) updateZosVersion(network string) error {
	substrateUrl := w.substrateUrl[network]

	substrateClient, err := newSubstrateClient(substrateUrl...)
	if err != nil {
		return err
	}

	currentZosVersion, err := substrateClient.checkVersion()
	if err != nil {
		return err
	}

	zosCurrent := fmt.Sprintf("%v/zos:%v.flist", w.src, currentZosVersion)

	zosLatest := fmt.Sprintf("%v:%v-3:latest.flist", w.dst, network)

	err = os.Symlink(zosCurrent, zosLatest)

	return err
}

func (w *worker) UpdateWithInterval() error {
	ticker := time.NewTicker(w.interval)
	quit := make(chan bool)
	var err error

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, network := range NETWORKS {
					err = w.updateZosVersion(network)
					if err != nil {
						quit <- true
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return err
}

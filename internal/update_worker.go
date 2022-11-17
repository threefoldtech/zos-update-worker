package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
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
	interval        time.Duration
	substrateUrl    map[string][]string
	substrateClient substrateClient

	logger zerolog.Logger
}

// new instance of the worker
func NewWorker(logger zerolog.Logger, src string, dst string, params map[string]interface{}) worker {
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
		logger:       logger,
	}
}

// setSubstrateClient sets the substrate client for a specific network
func (w *worker) setSubstrateClient(network string) error {
	w.logger.Debug().Msg("setting substrate client")

	substrateUrl := w.substrateUrl[network]

	substrateClient, err := newSubstrateClient(substrateUrl...)
	if err != nil {
		return err
	}

	w.substrateClient = substrateClient
	return nil
}

// check if a symlink exits
func symLinkExists(symLink string) error {
	if _, err := os.Lstat(symLink); err != nil {
		return err
	}

	return nil
}

// check if a symlink exits, remove it
func removeSymLinkIfExists(symLink string) error {
	if err := symLinkExists(symLink); err == nil {
		if err := os.Remove(symLink); err != nil {
			return fmt.Errorf("failed to unlink: %w", err)
		}
	}
	return nil
}

// updateZosVersion updates the latest zos flist for a specific network with the updated zos version
func (w *worker) updateZosVersion(network string) error {
	err := w.setSubstrateClient(network)
	if err != nil {
		return err
	}

	currentZosVersion, err := w.substrateClient.checkVersion()
	if err != nil {
		return err
	}

	w.logger.Debug().Msg(fmt.Sprintf("getting substrate version %v for network %v", currentZosVersion, network))

	zosCurrent := fmt.Sprintf("%v/zos:%v.flist", w.src, currentZosVersion)

	zosLatest := fmt.Sprintf("%v/zos:%v-3:latest.flist", w.dst, network)

	err = symLinkExists(zosCurrent)
	if err != nil {
		return err
	}

	err = removeSymLinkIfExists(zosLatest)
	if err != nil {
		return err
	}

	err = os.Symlink(zosCurrent, zosLatest)
	if err != nil {
		return err
	}

	w.logger.Debug().Msg(fmt.Sprintf("symlink %v to %v", zosCurrent, zosLatest))

	return nil
}

func (w *worker) UpdateWithInterval() error {
	ticker := time.NewTicker(w.interval)
	var err error

	for range ticker.C {
		for _, network := range NETWORKS {
			w.logger.Debug().Msg(fmt.Sprintf("updating zos version for %v", network))
			err = w.updateZosVersion(network)
			if err != nil {
				break
			}
		}

		if err != nil {
			ticker.Stop()
			break
		}
	}

	return err
}

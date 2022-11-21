package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	client "github.com/threefoldtech/substrate-client"
)

var SUBSTRATE_URLS = map[string][]string{
	"qa":         {"wss://tfchain.qa.grid.tf/ws"},
	"testing":    {"wss://tfchain.test.grid.tf/ws"},
	"production": {"wss://tfchain.grid.tf/ws"},
}

type Network string

var (
	MainNetwork Network = "production"
	TestNetwork Network = "testing"
	QANetwork   Network = "qa"
)

var NETWORKS = []Network{MainNetwork, QANetwork, TestNetwork}

type Params struct {
	Interval time.Duration
	QAUrls   []string
	TestUrls []string
	MainUrls []string
}

type worker struct {
	src string
	dst string

	params    Params
	substrate map[Network]client.Manager
}

// NewWorker creates a new instance of the worker
func NewWorker(src string, dst string, params Params) worker {
	var interval time.Duration

	if params.Interval == interval {
		params.Interval = time.Minute * 10
	}

	if len(params.QAUrls) == 0 {
		params.QAUrls = SUBSTRATE_URLS["qa"]
	}

	if len(params.TestUrls) == 0 {
		params.TestUrls = SUBSTRATE_URLS["testing"]
	}

	if len(params.MainUrls) == 0 {
		params.MainUrls = SUBSTRATE_URLS["production"]
	}

	substrate := map[Network]client.Manager{}
	substrate[MainNetwork] = client.NewManager(params.MainUrls...)
	substrate[QANetwork] = client.NewManager(params.QAUrls...)
	substrate[TestNetwork] = client.NewManager(params.TestUrls...)

	return worker{
		src:       src,
		dst:       dst,
		params:    params,
		substrate: substrate,
	}
}

// connection sets a new substrate connection
func (w *worker) connection(network Network) (*client.Substrate, error) {
	return w.substrate[network].Substrate()
}

// checkNetwork to check if a network is valid against main, qa, test
func checkNetwork(network Network) error {
	if network != MainNetwork && network != QANetwork && network != TestNetwork {
		return fmt.Errorf("invalid network")
	}

	return nil
}

// symLinkExists check if a symlink exits
func symLinkExists(symLink string) error {
	if _, err := os.Lstat(symLink); err != nil {
		return err
	}

	return nil
}

// removeSymLinkIfExists check if a symlink exits, remove it
func removeSymLinkIfExists(symLink string) error {
	if err := symLinkExists(symLink); err != nil {
		return err
	}

	if err := os.Remove(symLink); err != nil {
		return fmt.Errorf("failed to unlink: %w", err)
	}

	return nil
}

// updateZosVersion updates the latest zos flist for a specific network with the updated zos version
func (w *worker) updateZosVersion(network Network) error {
	if err := checkNetwork(network); err != nil {
		return err
	}

	con, err := w.connection(network)
	if err != nil {
		return err
	}
	defer con.Close()

	currentZosVersion, err := con.GetZosVersion()
	if err != nil {
		return err
	}

	log.Debug().Msg(fmt.Sprintf("getting substrate version %v for network %v", *currentZosVersion, network))

	zosCurrent := fmt.Sprintf("%v/zos:%v.flist", w.src, *currentZosVersion)
	zosLatest := fmt.Sprintf("%v/zos:%v-3:latest.flist", w.dst, network)

	err = symLinkExists(zosCurrent)
	if os.IsNotExist(err) {
		return err
	}

	if err == nil {
		log.Debug().Msg(fmt.Sprintf("symlink %v to %v already exists", zosCurrent, zosLatest))
		return nil
	}

	err = removeSymLinkIfExists(zosLatest)
	if err != nil {
		return err
	}

	err = os.Symlink(zosCurrent, zosLatest)
	if err != nil {
		return err
	}

	log.Debug().Msg(fmt.Sprintf("symlink %v to %v", zosCurrent, zosLatest))

	return nil
}

// UpdateWithInterval updates the latest zos flist for a specific network with the updated zos version
// with a specific interval between each update
func (w *worker) UpdateWithInterval() {
	ticker := time.NewTicker(w.params.Interval)

	for range ticker.C {
		for _, network := range NETWORKS {
			log.Debug().Msg(fmt.Sprintf("updating zos version for %v", network))

			exp := backoff.NewExponentialBackOff()
			exp.MaxInterval = 2 * time.Second
			exp.MaxElapsedTime = 30 * time.Second
			err := backoff.Retry(func() error {

				err := w.updateZosVersion(network)
				if err != nil {
					log.Error().Err(err).Msg("update failure. retrying")
				}
				return err

			}, exp)

			if err != nil {
				ticker.Stop()
				log.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			}
		}
	}
}

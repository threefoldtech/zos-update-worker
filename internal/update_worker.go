package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	client "github.com/threefoldtech/substrate-client"
)

type Network string

var (
	MainNetwork Network = "production"
	TestNetwork Network = "testing"
	QANetwork   Network = "qa"
)

type Params struct {
	Interval time.Duration
	QAUrls   []string
	TestUrls []string
	MainUrls []string
}

type worker struct {
	src string
	dst string

	interval  time.Duration
	substrate map[Network]client.Manager
}

// NewWorker creates a new instance of the worker
func NewWorker(src string, dst string, params Params) worker {
	substrate := map[Network]client.Manager{}

	if len(params.QAUrls) != 0 {
		substrate[QANetwork] = client.NewManager(params.QAUrls...)
	}

	if len(params.TestUrls) != 0 {
		substrate[TestNetwork] = client.NewManager(params.TestUrls...)
	}

	if len(params.MainUrls) != 0 {
		substrate[MainNetwork] = client.NewManager(params.MainUrls...)
	}

	return worker{
		src:       src,
		dst:       dst,
		substrate: substrate,
		interval:  params.Interval,
	}
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
func (w *worker) updateZosVersion(network Network, manager client.Manager) error {
	if err := checkNetwork(network); err != nil {
		return err
	}

	con, err := manager.Substrate()
	if err != nil {
		return err
	}
	defer con.Close()

	currentZosVersion, err := con.GetZosVersion()
	if err != nil {
		return err
	}

	log.Debug().Msgf("getting substrate version %v for network %v", currentZosVersion, network)

	zosCurrent := fmt.Sprintf("%v/zos:%v.flist", w.src, currentZosVersion)
	zosLatest := fmt.Sprintf("%v/zos:%v-3:latest.flist", w.dst, network)

	// check if current exists
	if _, err := os.Lstat(zosCurrent); err != nil {
		return err
	}

	// check if symlink exists
	dst, err := os.Readlink(zosLatest)

	if os.IsNotExist(err) {
		return err
	} else if err != nil {
		return err
	}

	// check if symlink is valid
	if dst == zosCurrent {
		log.Debug().Msgf("symlink %v to %v already exists", zosCurrent, zosLatest)
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

	log.Debug().Msgf("symlink %v to %v", zosCurrent, zosLatest)

	return nil
}

// UpdateWithInterval updates the latest zos flist for a specific network with the updated zos version
// with a specific interval between each update
func (w *worker) UpdateWithInterval() {
	ticker := time.NewTicker(w.interval)

	for range ticker.C {
		for network, manager := range w.substrate {
			log.Debug().Msgf("updating zos version for %v", network)

			exp := backoff.NewExponentialBackOff()
			exp.MaxInterval = 2 * time.Second
			exp.MaxElapsedTime = 10 * time.Second
			err := backoff.Retry(func() error {

				err := w.updateZosVersion(network, manager)
				if err != nil {
					log.Error().Err(err).Msg("update failure. retrying")
				}
				return err

			}, exp)

			if err != nil {
				log.Error().Err(err).Msg("update zos failed with error")
			}
		}
	}
}

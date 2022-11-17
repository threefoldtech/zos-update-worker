package internal

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestPkidStore(t *testing.T) {
	testDir := t.TempDir()
	logger := zerolog.New(os.Stdout).With().Logger()

	params := map[string]interface{}{}
	params["interval"] = 1 * time.Second
	src := testDir + "/tf-autobuilder"
	dst := testDir + "/tf-zos"

	err := os.Mkdir(src, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	err = os.Mkdir(dst, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	worker := NewWorker(logger, src, dst, params)

	t.Run("test_no_src_qa", func(t *testing.T) {
		err := worker.UpdateWithInterval()
		if err == nil {
			t.Errorf("update zos should fail")
		}
	})

	t.Run("test_no_src_test", func(t *testing.T) {
		_, err := os.Create(src + "/zos:v3.4.0-qa1.flist")
		if err != nil {
			t.Error(err)
		}

		err = worker.UpdateWithInterval()
		if err == nil {
			t.Errorf("update zos should fail for test, %v", err)
		}
	})

	t.Run("test_no_src_main", func(t *testing.T) {
		_, err = os.Create(src + "/zos:v3.1.1-rc2.flist")
		if err != nil {
			t.Error(err)
		}

		err = worker.UpdateWithInterval()
		if err == nil {
			t.Errorf("update zos should fail for main, %v", err)
		}
	})

	t.Run("test_params_wrong_url", func(t *testing.T) {
		params["qa"] = []string{"wss://tfchain.qa1.grid.tf/ws"}
		params["testing"] = []string{"wss://tfchain.test.grid.tf/ws"}
		params["production"] = []string{"wss://tfchain.grid.tf/ws"}

		worker = NewWorker(logger, src, dst, params)

		err := worker.UpdateWithInterval()
		if err == nil {
			t.Errorf("update zos should fail")
		}
	})
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rawdaGastan/zos-update-version/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zos-update-version",
	Short: "A worker to update the version of zos",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zerolog.New(os.Stdout).With().Logger()

		src, err := cmd.Flags().GetString("src")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}

		dst, err := cmd.Flags().GetString("dst")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}

		params := map[string]any{}
		interval, err := cmd.Flags().GetInt("interval")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}
		params["interval"] = time.Duration(interval) * time.Minute

		production, err := cmd.Flags().GetStringSlice("main-url")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}
		if len(production) > 0 {
			params["production"] = production
		}

		test, err := cmd.Flags().GetStringSlice("test-url")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}
		if len(test) > 0 {
			params["testing"] = test
		}

		qa, err := cmd.Flags().GetStringSlice("test-url")
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}
		if len(qa) > 0 {
			params["qa"] = qa
		}

		worker := internal.NewWorker(logger, src, dst, params)
		err = worker.UpdateWithInterval()
		if err != nil {
			logger.Error().Msg(fmt.Sprint("update zos failed with error: ", err))
			return
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()

	rootCmd.Flags().StringP("src", "s", "tf-autobuilder", "Enter your source directory")
	rootCmd.Flags().StringP("dst", "d", "tf-zos", "Enter your destination directory")
	rootCmd.Flags().IntP("interval", "i", 10, "Enter the interval between each update")

	rootCmd.Flags().StringSliceP("main-url", "m", []string{}, "Enter your mainnet substrate urls")
	rootCmd.Flags().StringSliceP("test-url", "t", []string{}, "Enter your testnet substrate urls")
	rootCmd.Flags().StringSliceP("qa-url", "q", []string{}, "Enter your qanet substrate urls")
}
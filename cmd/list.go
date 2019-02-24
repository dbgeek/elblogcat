// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/dbgeek/elblogcat/logworker"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print accesslogs that exists",
	Long: `Print all accesslogs that exists and possible to filter by

	* region
	* start time
	* loadbalancer id
`,
	Run: func(cmd *cobra.Command, args []string) {
		awsConfiguration := logworker.AWSconfiguration{Region: "eu-west-1"}
		configuration := logworker.NewConfiguration()
		accessLogFilter := logworker.NewAccessLogFilter()
		client := logworker.NewLogWorker(
			&awsConfiguration,
			&configuration,
			&accessLogFilter,
		)

		for _, v := range client.List() {
			fmt.Println(v)
		}

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

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

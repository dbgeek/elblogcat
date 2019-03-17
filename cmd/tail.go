package cmd

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dbgeek/elblogcat/logcat"
	"github.com/dbgeek/elblogcat/logworker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tailCmd represents the tail command
var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Porman tail pool for new accesslogs for default every 1min",
	Long: `
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
		logs := make(chan string, 1)

		client.Tail(logs)

		for v := range logs {
			buff := &aws.WriteAtBuffer{}
			key := fmt.Sprintf("%s%s", accessLogFilter.AccesslogPath(configuration.Prefix), v)
			_, err := client.S3Downloader.Download(buff, &s3.GetObjectInput{
				Bucket: aws.String(viper.GetString("s3-bucket")),
				Key:    aws.String(key),
			})
			if err != nil {
				logworker.Logger.Fatalf("Failed to Download key: %v from s3. Got error: %v",
					key,
					err)
			}

			c := logcat.NewRowFilter()
			b := bytes.NewBuffer(buff.Bytes())
			a := logcat.Accesslog{
				Content:     b,
				RowFilter:   c,
				PrintFields: viper.GetString("fields"),
			}
			a.Cat()
		}
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)
	tailCmd.PersistentFlags().Duration("polling-interval", 60*time.Second, "")
	viper.BindPFlag("polling-interval", tailCmd.PersistentFlags().Lookup("polling-interval"))

}

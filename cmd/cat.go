package cmd

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dbgeek/elblogcat/logcat"
	"github.com/dbgeek/elblogcat/logworker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "cat accesslog from s3",
	Long: `Download the accesslog and cat it

possible user these filter
* client-ip
* elb-status-code
* target-status-code
* http-method
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
			buff := &aws.WriteAtBuffer{}
			key := fmt.Sprintf("%s%s%s", configuration.Prefix, accessLogFilter.AccesslogPath(), v)
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
				Content:   b,
				RowFilter: c,
			}
			for _, row := range a.Cat() {
				fmt.Println(row)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(catCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// catCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//catCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	catCmd.PersistentFlags().StringP("client-ip", "", ".*", "")
	viper.BindPFlag("client-ip", catCmd.PersistentFlags().Lookup("client-ip"))
	catCmd.PersistentFlags().StringP("elb-status-code", "", ".*", "")
	viper.BindPFlag("elb-status-code", catCmd.PersistentFlags().Lookup("elb-status-code"))
	catCmd.PersistentFlags().StringP("target-status-code", "", ".*", "")
	viper.BindPFlag("target-status-code", catCmd.PersistentFlags().Lookup("target-status-code"))
	catCmd.PersistentFlags().StringP("http-method", "", ".*", "")
	viper.BindPFlag("http-method", catCmd.PersistentFlags().Lookup("http-method"))
}

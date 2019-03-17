package cmd

import (
	"fmt"
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "elblogcat",
	Short: "List & cat aws lb acesses",
	Long: `List & cat aws alb/elb accesslog that are stored in s3.
	
Filter output is possible by:
* timerange for one day
* loadbalancer id
* loadbalancer ip-address
* accesslog unique string
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.elblogcat.yaml)")
	rootCmd.PersistentFlags().StringP("aws-account-id", "a", "", "The AWS account ID of the owner.")
	viper.BindPFlag("aws-account-id", rootCmd.PersistentFlags().Lookup("aws-account-id"))
	rootCmd.PersistentFlags().StringP("region", "r", "", "The region for your load balancer and S3 bucket.")
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	rootCmd.PersistentFlags().StringP("load-balancer-id", "l", ".*", "The resource ID of the load balancer. If the resource ID contains any forward slashes (/), they are replaced with periods (.).")
	viper.BindPFlag("load-balancer-id", rootCmd.PersistentFlags().Lookup("load-balancer-id"))
	rootCmd.PersistentFlags().StringP("ip-address", "i", ".*", "The IP address of the load balancer node that handled the request. For an internal load balancer, this is a private IP address.")
	viper.BindPFlag("ip-address", rootCmd.PersistentFlags().Lookup("ip-address"))
	rootCmd.PersistentFlags().StringP("random-string", "s", ".*", "A system-generated random string.")
	viper.BindPFlag("random-string", rootCmd.PersistentFlags().Lookup("random-string"))
	rootCmd.PersistentFlags().StringP("s3-bucket", "b", ".*", "The name of the S3 bucket.")
	viper.BindPFlag("s3-bucket", rootCmd.PersistentFlags().Lookup("s3-bucket"))
	rootCmd.PersistentFlags().StringP("s3-prefix", "p", ".*", "The prefix (logical hierarchy) in the bucket. If you don't specify a prefix, the logs are placed at the root level of the bucket.")
	viper.BindPFlag("s3-prefix", rootCmd.PersistentFlags().Lookup("s3-prefix"))
	rootCmd.PersistentFlags().StringP("start-time", "", defaultStartTime().Format("2006-01-02 15:04:05"), "")
	viper.BindPFlag("start-time", rootCmd.PersistentFlags().Lookup("start-time"))
	rootCmd.PersistentFlags().StringP("end-time", "", time.Now().Format("2006-01-02 15:04:05"), "")
	viper.BindPFlag("end-time", rootCmd.PersistentFlags().Lookup("end-time"))
	rootCmd.PersistentFlags().Int64P("max-keys", "", 500, "control nr of keys that should be return from s3 api for each response.")
	viper.BindPFlag("max-keys", rootCmd.PersistentFlags().Lookup("max-keys"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".elblogcat" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".elblogcat")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func defaultStartTime() time.Time {
	t := time.Now().UTC()
	d := 24 * time.Hour
	return t.Truncate(d)
}

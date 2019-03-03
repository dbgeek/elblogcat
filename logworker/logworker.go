package logworker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	LogWorker struct {
		Config          *AWSconfiguration
		S3              *s3.S3
		S3Downloader    *s3manager.Downloader
		Configuration   *Configuration
		AccessLogFilter *AccessLogFilter
	}
	//AWSconfiguration --..
	AWSconfiguration struct {
		Region  string
		Profile string
	}
	Configuration struct {
		Bucket string
		Prefix string
	}
	AccessLogFilter struct {
		matchString    string
		AwsAccountID   string
		Region         string
		LoadBalancerID string
		IPaddress      string
		RandomString   string
		StartTime      time.Time
		EndTime        time.Time
		matcher        *regexp.Regexp
	}
)

var (
	Logger *logrus.Logger
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	Logger.SetOutput(os.Stdout)

}

func newFilter(accessLogFilter *AccessLogFilter) *regexp.Regexp {
	matchString := fmt.Sprintf("^(%s)_(elasticloadbalancing)_(%s)_(%s)_(%s)_(%s)_(%s).log.gz$",
		accessLogFilter.AwsAccountID,
		accessLogFilter.Region,
		accessLogFilter.LoadBalancerID,
		".*",
		accessLogFilter.IPaddress,
		accessLogFilter.RandomString,
	)

	regexp, err := regexp.Compile(matchString)
	if err != nil {
		Logger.Fatalf("Failed to compile matchstring. Gott error: %v", err)
	}

	return regexp
}

func NewLogWorker(
	awsConfiguration *AWSconfiguration,
	configuration *Configuration,
	accessLogFilter *AccessLogFilter,
) *LogWorker {
	logWorker := LogWorker{}
	logWorker.Configuration = configuration
	logWorker.AccessLogFilter = accessLogFilter
	logWorker.AccessLogFilter.matcher = newFilter(accessLogFilter)

	awsCfg := aws.Config{}
	if awsConfiguration.Region != "" {
		awsCfg.Region = &awsConfiguration.Region
	}

	awsSessionOpts := session.Options{
		Config:                  awsCfg,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	}

	if awsConfiguration.Profile != "" {
		awsSessionOpts.Profile = awsConfiguration.Profile
	}

	sess := session.Must(session.NewSessionWithOptions(awsSessionOpts))

	logWorker.S3 = s3.New(sess)
	logWorker.S3Downloader = s3manager.NewDownloader(sess)

	return &logWorker
}

func (l *LogWorker) List() []string {

	var accessLogs []string
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(l.Configuration.Bucket),
		Prefix:    aws.String(l.AccessLogFilter.AccesslogPath(l.Configuration.Prefix)),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(200),
	}
	err := l.S3.ListObjectsV2Pages(input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, val := range page.Contents {
				accessLog := strings.Split(*val.Key, "/")[len(strings.Split(*val.Key, "/"))-1]
				if l.AccessLogFilter.matcher.MatchString(accessLog) {
					accessLogs = append(accessLogs, accessLog)
				}
			}
			return true
		})
	if err != nil {
		fmt.Println(err)
	}
	return accessLogs
}

func (a *AccessLogFilter) AccesslogPath(prefix string) string {
	return filepath.Join(prefix, fmt.Sprintf("AWSLogs/%s/elasticloadbalancing/%s/%s/", a.AwsAccountID, a.Region, a.StartTime.Format("2006/01/02"))) + "/"

}

func (a *AccessLogFilter) filterByTime(accessLog string) bool {
	accessLogEndTimeStr := strings.Split(accessLog, "_")[4]
	accessLogEndTimeStamp, err := time.Parse("20060102T1504Z", accessLogEndTimeStr)
	if err != nil {
		Logger.Fatalf("failed to parse timestamp for accesslog name")
	}
	accessLogStartTimeStamp := accessLogEndTimeStamp.Add(-5 * time.Minute)

	if (a.StartTime.Before(accessLogStartTimeStamp) || a.StartTime == accessLogStartTimeStamp) &&
		(a.EndTime.After(accessLogEndTimeStamp) || accessLogEndTimeStamp == a.EndTime) {
		Logger.Debugf("1 if. aEndTimeStamp: %v, endFilter: %v \n", accessLogEndTimeStamp.Format("15:04"), a.EndTime.Format("15:04"))
		return true
	} else if (a.StartTime.After(accessLogStartTimeStamp) && a.StartTime.Before(accessLogEndTimeStamp)) &&
		(a.EndTime.Before(accessLogEndTimeStamp) && a.EndTime.After(a.StartTime)) {
		Logger.Debugln("2 if")
		return true
	} else if (a.StartTime.Before(accessLogStartTimeStamp) || a.StartTime == accessLogStartTimeStamp) &&
		(a.EndTime.Before(accessLogEndTimeStamp) && a.EndTime.After(a.StartTime) && a.EndTime.After(accessLogStartTimeStamp)) {
		Logger.Debugln("3 if")
		return true
	} else if (a.EndTime.After(accessLogEndTimeStamp) || a.EndTime == accessLogEndTimeStamp) &&
		(a.StartTime.After(accessLogStartTimeStamp) && a.StartTime.Before(accessLogEndTimeStamp)) {
		Logger.Debugln("4 if")
		return true
	}
	return false
}

func NewAccessLogFilter() AccessLogFilter {

	startTime, err := time.Parse("2006-01-02 15:04:05", viper.GetString("start-time"))
	if err != nil {
		Logger.Fatalf("Failed to parse start time. Gott error: %v", err)
		fmt.Println("failed to parse starttime")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", viper.GetString("end-time"))
	if err != nil {
		Logger.Fatalf("Failed to parse end time. Gott error: %v", err)
		fmt.Println("failed to parse endtime")
	}
	accessLogFilter := AccessLogFilter{}
	accessLogFilter.AwsAccountID = viper.GetString("aws-account-id")
	accessLogFilter.Region = viper.GetString("region")
	accessLogFilter.StartTime = startTime // time.Now()
	accessLogFilter.EndTime = endTime
	accessLogFilter.LoadBalancerID = viper.GetString("load-balancer-id")
	accessLogFilter.IPaddress = viper.GetString("ip-address")
	accessLogFilter.RandomString = viper.GetString("random-string")

	return accessLogFilter
}

func NewConfiguration() Configuration {
	return Configuration{
		Bucket: viper.GetString("s3-bucket"),
		Prefix: viper.GetString("s3-prefix"),
	}
}

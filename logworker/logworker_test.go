package logworker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAccessLogFilterAccesslogPath(t *testing.T) {

	tt := []struct {
		name            string
		accessLogFilter AccessLogFilter
		prefix          string
		out             string
	}{
		{"WithoutPrefix",
			AccessLogFilter{

				AwsAccountID: "00000000111111",
				Region:       "eu-west-1",
				StartTime:    time.Now(),
			},
			"",
			fmt.Sprintf("AWSLogs/00000000111111/elasticloadbalancing/eu-west-1/%s/", time.Now().Format("2006/01/02")),
		},
		{"WithPrefix",
			AccessLogFilter{

				AwsAccountID: "00000000111111",
				Region:       "eu-west-1",
				StartTime:    time.Now(),
			},
			"team-xxx",
			fmt.Sprintf("team-xxx/AWSLogs/00000000111111/elasticloadbalancing/eu-west-1/%s/", time.Now().Format("2006/01/02")),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.accessLogFilter.AccesslogPath(tc.prefix) != tc.out {
				t.Fatalf("accespath test %v should be %v; got %v", tc.name, tc.out, tc.accessLogFilter.AccesslogPath(tc.prefix))
			}
		})
	}

}

func TestMatcher(t *testing.T) {
	tt := []struct {
		name           string
		AwsAccountID   string
		Region         string
		LoadBalancerID string
		IPaddress      string
		RandomString   string
		in             string
	}{
		{
			"match-everything",
			".*",
			".*",
			".*",
			".*",
			".*",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1610Z_10.205.19.102_34cjbbr9.log.gz",
		},
		{
			"match-accountid",
			"0123456789",
			".*",
			".*",
			".*",
			".*",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1610Z_10.205.19.102_34cjbbr9.log.gz",
		},
		{
			"match-loadbalancerid",
			".*",
			".*",
			"elb-prod.32435435",
			".*",
			".*",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1610Z_10.205.19.102_34cjbbr9.log.gz",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			accessLogFilter := AccessLogFilter{}
			accessLogFilter.AwsAccountID = tc.AwsAccountID
			accessLogFilter.Region = tc.Region
			accessLogFilter.LoadBalancerID = tc.LoadBalancerID
			accessLogFilter.IPaddress = tc.IPaddress
			accessLogFilter.RandomString = tc.RandomString
			matcher := newFilter(&accessLogFilter)
			if !matcher.MatchString(tc.in) {
				t.Fatalf("")
			}
		})
	}
}

func TestStartTimeEndTime(t *testing.T) {
	tt := []struct {
		name      string
		StartTime string
		EndTime   string
		in        string
		out       bool
	}{
		{
			"FalseStartEndFilterAfterAccessLogTimestamp",
			"2019-02-23 15:00:00",
			"2019-02-23 16:10:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1110Z_10.205.19.102_34cjbbr9.log.gz",
			false,
		},
		{
			"FalseStartEndFilterBeforeAccesslogTimestamå",
			"2019-02-23 09:00:00",
			"2019-02-23 10:10:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1110Z_10.205.19.102_34cjbbr9.log.gz",
			false,
		},
		{
			"TrueBetweenStartEnd",
			"2019-02-23 15:00:00",
			"2019-02-23 23:00:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1610Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
		{
			"TrueSameEndTime",
			"2019-02-23 15:00:00",
			"2019-02-23 16:10:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1610Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
		{
			"TrueSameStartTime",
			"2019-02-23 14:50:00",
			"2019-02-23 16:10:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1455Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
		{
			"TrueStartAndEndTimeBetweenAcessLogTimestamp",
			"2019-02-23 14:52:00",
			"2019-02-23 14:54:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1455Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
		{
			"TrueStartFilterAfterAcessLogFirstTimestamp",
			"2019-02-23 14:52:00",
			"2019-02-23 16:00:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1455Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
		{
			"TrueEndFilterBeforeAcessLogEndTimestamp",
			"2019-02-23 14:45:00",
			"2019-02-23 14:54:00",
			"0123456789_elasticloadbalancing_eu-west-1_elb-prod.32435435_20190223T1455Z_10.205.19.102_34cjbbr9.log.gz",
			true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sTime, _ := time.Parse("2006-01-02 15:04:05", tc.StartTime)
			eTime, _ := time.Parse("2006-01-02 15:04:05", tc.EndTime)
			accessLogFilter := AccessLogFilter{}
			accessLogFilter.StartTime = sTime
			accessLogFilter.EndTime = eTime

			if !accessLogFilter.filterByTime(tc.in) == tc.out {
				t.Fatalf("startTime: %v, endTime: %v, accesslog timestamp: %v",
					sTime.Format("20060102T15:04Z"),
					eTime.Format("20060102T15:04Z"),
					strings.Split(tc.in, "_")[4])
			}
		})
	}
}

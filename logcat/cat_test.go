package logcat

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func TestAccessLogFilter(t *testing.T) {
	tt := []struct {
		name             string
		testData         []byte
		clientIP         string
		ElbStatusCode    string
		targetStatusCode string
		HTTPmethod       string
		out              bool
	}{
		{
			"no-filter",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.222.161.42:32774 10.222.20.10:443 0.000 0.002 0.000 200 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			".*",
			".*",
			".*",
			".*",
			true,
		},
		{
			"filter-clientIP",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.222.161.42:32774 10.222.20.10:443 0.000 0.002 0.000 200 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			"10.222.161.42",
			".*",
			".*",
			".*",
			true,
		},
		{
			"filter-elbstatuscode",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.222.161.42:32774 10.222.20.10:443 0.000 0.002 0.000 200 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			".*",
			"200",
			".*",
			".*",
			true,
		},
		{
			"filter-elbstatuscode-startwith-2",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.222.161.42:32774 10.222.20.10:443 0.000 0.002 0.000 250 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			".*",
			"2.*",
			".*",
			".*",
			true,
		},
		{
			"filter-targetstatuscode",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.222.161.42:32774 10.222.20.10:443 0.000 0.002 0.000 200 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			".*",
			".*",
			"200",
			".*",
			true,
		},
		{
			"filter-elbstatuscode",
			[]byte(`https 2019-02-02T00:14:07.437021Z elb01 10.225.161.42:32774 10.225.24.10:443 0.000 0.002 0.000 200 200 371 178 "GET https://elb01.prod.com:443/status HTTP/1.1" "Faraday v0.9.2" ECDHE-RSA-AES128-GCM-SHA256 TLSv1.2 arn:aws:elasticloadbalancing:eu-west-1:0123456789:targetgroup/prod-tg/8f858d88ba9c836c "Root=1-xxxxx-yyyyyyyyyyyyyyyyyyyyy" "elb01.prod.com" "arn:aws:acm:eu-west-1:0123456789:certificate/bbbbbbbb-1cbf-4f99-aaaa-cccccccccccc" 0 2019-02-02T00:14:07.435000Z "forward" "-" "-"`),
			".*",
			".*",
			".*",
			"GET",
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b := []byte{}
			buff := bytes.NewBuffer(b)
			gw := gzip.NewWriter(buff)
			gw.Write(tc.testData)
			gw.Close()
			a := Accesslog{}
			a.Content = buff
			a.RowFilter = Filter{
				ClientIP:         tc.clientIP,
				HTTPmethod:       tc.HTTPmethod,
				ElbStatusCode:    tc.ElbStatusCode,
				TargetStatusCode: tc.targetStatusCode,
			}
			result := a.Cat()
			if len(result) == 0 {
				t.Fatalf("test: %s failed to find match", tc.name)
			}

		})
	}

}

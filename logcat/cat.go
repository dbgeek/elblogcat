package logcat

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/dbgeek/elblogcat/logworker"
	"github.com/spf13/viper"
)

type (
	Accesslog struct {
		Content     *bytes.Buffer
		RowFilter   Filter
		PrintFields string
	}
	Filter struct {
		ClientIP         string
		ElbStatusCode    string
		TargetStatusCode string
		HTTPmethod       string
	}

	entry struct {
		row string
	}
	rowMatch struct {
		matchString string
		matcher     *regexp.Regexp
	}
	accessLogFieldPosition map[string]int
)

var (
	fieldPosition = map[string]int{
		"conn-type":                0,
		"timestamp":                1,
		"elb":                      2,
		"client:port":              3,
		"target:port":              4,
		"request_porecessing_time": 5,
		"target_processing_time":   6,
		"respsone_processing_time": 7,
		"elbStatis_code":           8,
		"targetStatus_code":        9,
		"received_bytes":           10,
		"send_bytes":               11,
		"request":                  12,
		"user-agent":               13,
		"ssl-cipher":               14,
		"ssl-protocol":             15,
		"target-group-arn":         16,
		"trace_id":                 17,
		"domain_name":              18,
		"chose_cert_arn":           19,
		"marched_rule_priority":    20,
		"request_creation_time":    21,
		"action_executed":          22,
		"redirect_url":             23,
		"error_reason":             24,
	}
)

func (a *Accesslog) Cat() {
	//ToDO Break out this to it own method on the acceslog
	var pFields []int
	for _, v := range strings.Fields(a.PrintFields) {
		pFields = append(pFields, fieldPosition[v])
	}

	gzReader, err := gzip.NewReader(a.Content)
	if err != nil {
		logworker.Logger.Fatalf("new gzip reader failed with: %v", err)
	}
	scanner := bufio.NewScanner(gzReader)
	filter := newRowMatch(a.RowFilter)
	var entries [][]string
	for scanner.Scan() {
		if filter.matcher.MatchString(scanner.Text()) {
			//This should be break out to it´s own function
			r := csv.NewReader(strings.NewReader(scanner.Text()))
			r.Comma = ' '
			field, err := r.Read()
			if err != nil {
				fmt.Println(err)
			}
			entries = append(entries, field)
		}
	}
	tw := new(tabwriter.Writer)
	tw.Init(os.Stdout, 0, 8, 2, '\t', 0)
	defer tw.Flush()
	for _, v := range entries {
		var str string
		for _, val := range pFields {
			str += fmt.Sprintf("%s\t", v[val])
		}
		fmt.Fprintln(tw, str)
	}

}
func newRowMatch(filter Filter) *rowMatch {
	r := rowMatch{}
	r.matchString = fmt.Sprintf("^(.*) (.*) (.*) (%s:.*) (.*) (.*) (.*) (.*) (%s) (%s) (.*) (.*) (\"%s.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*) (.*)$",
		filter.ClientIP,
		filter.ElbStatusCode,
		filter.TargetStatusCode,
		filter.HTTPmethod)

	regExp, err := regexp.Compile(r.matchString)
	if err != nil {
		logworker.Logger.Fatalf("Failed rowmatch  compile regexp got error: %v", err)
	}
	r.matcher = regExp
	return &r
}

func NewRowFilter() Filter {
	return Filter{
		ClientIP:         viper.GetString("client-ip"),
		ElbStatusCode:    viper.GetString("elb-status-code"),
		TargetStatusCode: viper.GetString("target-status-code"),
		HTTPmethod:       viper.GetString("http-method"),
	}
}

package logcat

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"regexp"

	"github.com/dbgeek/elblogcat/logworker"
	"github.com/spf13/viper"
)

type (
	Accesslog struct {
		Content   *bytes.Buffer
		RowFilter Filter
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
)

func (a *Accesslog) Cat() []string {
	gzReader, err := gzip.NewReader(a.Content)
	if err != nil {
		logworker.Logger.Fatalf("new gzip reader failed with: %v", err)
	}
	s := []string{}
	scanner := bufio.NewScanner(gzReader)
	filter := newRowMatch(a.RowFilter)
	for scanner.Scan() {
		if filter.matcher.MatchString(scanner.Text()) {
			s = append(s, scanner.Text())
		}
	}
	return s

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

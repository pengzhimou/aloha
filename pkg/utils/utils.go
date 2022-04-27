package utils

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"os"
	"regexp"
	"time"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var startTime = time.Now()

// now returns time.Duration using stdlib time
func Now() time.Duration { return time.Since(startTime) }

func CloneRequest(r *http.Request, body []byte) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	if len(body) > 0 {
		r2.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return r2
}

func ParseURL(urlstring string) *gourl.URL {
	if urlstring == "" {
		return nil
	} else {
		proxyURL, err := gourl.Parse(urlstring)
		if err != nil {
			ErrAndExit(err.Error())
			return nil
		} else {
			return proxyURL
		}
	}
}

func ErrAndExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func UsageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func ParseInputWithRegexp(input, regx string) ([]string, error) {
	re := regexp.MustCompile(regx)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse the provided input; input = %v", input)
	}
	return matches, nil
}

func MarshalS(x interface{}) string {
	s, _ := json.Marshal(x)
	return string(s)
}

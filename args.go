package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"time"
)

type Args struct {
	// client
	DisableCompression bool
	DisableKeepAlives  bool
	DisableRedirects   bool
	ProxyAddr          string
	Certfile           string
	Keyfile            string
	H2                 bool

	// req
	Url         string
	Method      string
	HeaderSlice []string
	Payload     string
	Accept      string
	ContentType string
	AuthHeader  string
	HostHeader  string
	UserAgent   string
	UrlFile     string
	PayloadFile string

	// task
	Conc       int
	Nums       int
	Qps        float64
	Timeout    int
	Duration   time.Duration
	Round      int
	Roundsleep int
	Cpus       int
	Output     string
}

type HTTPClient struct {
}

type HTTPReq struct {
}

type HTTPResp struct {
}

func NewArgs() *Args {
	return &Args{}
}

func (a *Args) Init() {
	url = flag.String("url", "", "")                // request
	method = flag.String("m", "GET", "")            // request
	var hs headerSlice                              // request
	flag.Var(&hs, "H", "")                          // request
	payload = flag.String("pl", "", "")             // request
	accept = flag.String("A", "", "")               // header
	contentType = flag.String("T", "text/html", "") // header
	authHeader = flag.String("a", "", "")           // header
	hostHeader = flag.String("host", "", "")        // header
	userAgent = flag.String("U", "", "")            // header
	urlFile = flag.String("urlfile", "", "")        // request
	payloadFile = flag.String("plfile", "", "")     // request

	disableCompression = flag.Bool("disable-compression", false, "") // tr
	disableKeepAlives = flag.Bool("disable-keepalive", false, "")    // tr
	disableRedirects = flag.Bool("disable-redirects", false, "")     // tr
	proxyAddr = flag.String("x", "", "")                             // tr
	certfile = flag.String("cert", "", "")                           // tr
	keyfile = flag.String("key", "", "")                             // tr
	h2 = flag.Bool("h2", false, "")                                  // tr

	conc = flag.Int("c", 50, "")                        // task
	nums = flag.Int("n", 200, "")                       // task
	qps = flag.Float64("q", 0, "")                      // task
	timeout = flag.Int("t", 20, "")                     // task
	duration = flag.Duration("z", 0, "")                // task
	round = flag.Int("r", 1, "")                        // task
	roundsleep = flag.Int("rs", 0, "")                  // task
	cpus = flag.Int("cpus", runtime.GOMAXPROCS(-1), "") // task
	output = flag.String("o", "", "")                   // task

	flag.Parse()

	if flag.NFlag() < 1 {
		usageAndExit("")
	}

}

// func userKill(w *requester.Work) {
// 	// 处理用户终止ctrl-c，调用stop
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt)
// 	go func() {
// 		<-c
// 		w.Stop()
// 	}()
// }

func errAndExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func parseInputWithRegexp(input, regx string) ([]string, error) {
	re := regexp.MustCompile(regx)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse the provided input; input = %v", input)
	}
	return matches, nil
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

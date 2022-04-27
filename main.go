package main

import (
	"aloha/pkg/tasks"
	"aloha/pkg/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
)

var usage = `Usage: aloha [options...]

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of workers to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in queries per second (QPS) per worker. Default is no rate limit.
  -z  Duration of application to send requests. When duration is reached,
      application stops and exits. If duration is specified, n is ignored.
      Examples: -z 10s -z 3m.
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -U  User-Agent, defaults to version "aloha/0.0.1".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -disable-redirects    Disable following of HTTP redirects
  -cpus                 Number of used cpu cores.
                        (default for current machine is %d cores)

  -cert certfile location
  -key keyfile location
  -urlfile urlfile location
  -url url link
  -r rounds
  -rs each round skip time
  -streamfile stream file location
`

func main() {
	Url := flag.String("url", "", "")     // request
	Method := flag.String("m", "GET", "") // request
	HeaderSlc := tasks.HeaderSlice{}
	flag.Var(&HeaderSlc, "H", "")                    // request
	Payload := flag.String("pl", "", "")             // request
	Accept := flag.String("A", "", "")               // header
	ContentType := flag.String("T", "text/html", "") // header
	AuthHeader := flag.String("a", "", "")           // header
	HostHeader := flag.String("host", "", "")        // header
	UserAgent := flag.String("U", "", "")            // header
	UrlFile := flag.String("urlfile", "", "")        // request
	PayloadFile := flag.String("plfile", "", "")     // request

	DisableCompression := flag.Bool("disable-compression", false, "") // tr
	DisableKeepAlives := flag.Bool("disable-keepalive", false, "")    // tr
	DisableRedirects := flag.Bool("disable-redirects", false, "")     // tr
	ProxyAddr := flag.String("x", "", "")                             // tr
	Certfile := flag.String("cert", "", "")                           // tr
	Keyfile := flag.String("key", "", "")                             // tr
	H2 := flag.Bool("h2", false, "")                                  // tr

	Conc := flag.Int("c", 50, "")                        // task
	Nums := flag.Int("n", 200, "")                       // task
	Qps := flag.Float64("q", 0, "")                      // task
	Timeout := flag.Int("t", 20, "")                     // task
	Duration := flag.Duration("z", 0, "")                // task
	Round := flag.Int("r", 1, "")                        // task
	Roundsleep := flag.Int("rs", 0, "")                  // task
	Cpus := flag.Int("cpus", runtime.GOMAXPROCS(-1), "") // task
	Output := flag.String("o", "", "")                   // task

	StreamFile := flag.String("streamfile", "", "") // 编排

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()

	if flag.NFlag() < 1 {
		utils.UsageAndExit("")
	}

	args := tasks.NewArgs(
		&tasks.ClientArgs{
			DisableCompression: *DisableCompression,
			DisableKeepAlives:  *DisableKeepAlives,
			DisableRedirects:   *DisableRedirects,
			ProxyAddr:          *ProxyAddr,
			Certfile:           *Certfile,
			Keyfile:            *Keyfile,
			H2:                 *H2,
		},
		&tasks.ReqArgs{
			Url:         *Url,
			Method:      *Method,
			HeaderSlc:   HeaderSlc,
			Payload:     *Payload,
			Accept:      *Accept,
			ContentType: *ContentType,
			AuthHeader:  *AuthHeader,
			HostHeader:  *HostHeader,
			UserAgent:   *UserAgent,
			UrlFile:     *UrlFile,
			PayloadFile: *PayloadFile,
			StreamFile:  *StreamFile,
		},
		&tasks.TaskArgs{
			Conc:       *Conc,
			Nums:       *Nums,
			Qps:        *Qps,
			Timeout:    *Timeout,
			Duration:   *Duration,
			Round:      *Round,
			Roundsleep: *Roundsleep,
			Cpus:       *Cpus,
			Output:     *Output,
		},
	)

	fmt.Println(utils.MarshalS(args), 111111111)

	task := args.GenReqArgsForSingleUrlTask()
	fmt.Println(utils.MarshalS(task), 2222222)

	// task.NewClient(args)
	// fmt.Println(*task.Clt, 33333333)

	// for _, job := range task.ReqOpts {
	// 	job.NewRequest()
	// 	fmt.Println(*job.Req, 444444444)

	// 	resp, err := task.Clt.Do(job.Req)

	// 	fmt.Println(err, "errrr11111111111")

	// 	bodybyte, err := ioutil.ReadAll(resp.Body)
	// 	fmt.Println(err, "errrr22222222222")
	// 	fmt.Println(string(bodybyte))
	// }

	/* 复用req需要解决复用问题，后面再弄
	task.NewClient(args)
	task.NewRequest()

	for _, job := range task.ReqOpts {
		job.GenRequest(&task)

		resp, err := task.Clt.Do(job.Req)

		fmt.Println(err, "errrr11111111111")

		bodybyte, err := ioutil.ReadAll(resp.Body)
		fmt.Println(err, "errrr22222222222")
		fmt.Println(string(bodybyte))
	}
	*/

}

package tasks

import (
	"aloha/pkg/utils"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

type Args struct {
	CArgs ClientArgs
	RArgs ReqArgs
	TArgs TaskArgs
}
type ClientArgs struct {
	DisableCompression bool
	DisableKeepAlives  bool
	DisableRedirects   bool
	ProxyAddr          string
	Certfile           string
	Keyfile            string
	H2                 bool
}

type ReqArgs struct {
	Url         string
	Methord     string
	HeaderSlc   HeaderSlice
	Payload     string
	Accept      string
	ContentType string
	AuthHeader  string
	HostHeader  string
	UserAgent   string
	UrlFile     string
	PayloadFile string
	StreamFile  string
}

type TaskArgs struct {
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

func NewArgs(cltArgs *ClientArgs, reqArgs *ReqArgs, taskArgs *TaskArgs) *Args {
	return &Args{
		CArgs: *cltArgs,
		RArgs: *reqArgs,
		TArgs: *taskArgs,
	}
}

func (a *Args) BasicAuth() (username, password *string) {
	// set basic auth if set
	if a.RArgs.AuthHeader != "" {
		match, err := utils.ParseInputWithRegexp(a.RArgs.AuthHeader, authRegexp)
		if err != nil {
			utils.UsageAndExit(err.Error())
		}
		username, password = &match[1], &match[2]
	}
	return
}

func (a *Args) RunnerInit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	brk := false
	for r := 0; r <= a.TArgs.Round; r++ {

		if r != a.TArgs.Round {
			select {
			case <-c:
				brk = true
				break
			default:
				go a.TasksGen()
				if a.TArgs.Round > 1 {
					fmt.Printf("Finished Round: %v, start to sleep:%v second\n", r+1, a.TArgs.Roundsleep)
					fmt.Println("---------------------------------")
				}
			}
			if brk {
				break
			}
			time.Sleep(time.Duration(a.TArgs.Roundsleep) * time.Second)
		}
	}
}

func (a *Args) TasksGen() {

	alltask := NewTaskFull()
	alltask.NewClient(a)

	wg := sync.WaitGroup{}
	if a.RArgs.UrlFile == "" {
		wg.Add(1)
		taskfull := NewTaskFull()
		taskfull.Init()
		taskfull.Run(taskfull.Clt)

		taskfull.NewClient(a)

		task := a.GenReqArgsForSingleTask(a.RArgs.Url)
		taskfull.Task = task

		wg.Wait()
	} else if a.RArgs.UrlFile != "" {
		// data, err := ioutil.ReadFile(a.RArgs.UrlFile)
		// if err != nil {
		// 	utils.ErrAndExit(fmt.Sprintf("---read fail: %s", err.Error()))
		// }
		// for _, line := range strings.Split(string(data), "\n") {
		// 	if !strings.Contains(line, "http") { //处理空行和换行符
		// 		continue
		// 	}
		// 	wg.Add(1)
		// 	task := a.GenReqArgsForSingleTask(line)
		// 	task.Run(alltask.Clt)
		// }
		// wg.Wait()
	} else if a.RArgs.StreamFile != "" {
		// later
		fmt.Println("Empty Logic, Later..")
	}
}

func (a *Args) GenReqArgsForSingleTask(url string) SingleTask {
	hrall := SingleTask{}
	hr := &ReqOpt{}

	// url部分========
	hr.Url = &url

	// methord部分
	hr.Methord = &a.RArgs.Methord

	// header部分=========
	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", a.RArgs.ContentType)

	for _, h := range a.RArgs.HeaderSlc {
		match, err := utils.ParseInputWithRegexp(h, headerRegexp)
		if err != nil {
			utils.UsageAndExit(err.Error())
		}
		header.Set(match[1], match[2])
	}

	if a.RArgs.Accept != "" {
		header.Set("Accept", a.RArgs.Accept)
	}

	if a.RArgs.UserAgent != "" {
		header.Set("User-Agent", a.RArgs.UserAgent+" "+heyUA)
	}

	hr.Header = &header

	// payload部分=========
	if a.RArgs.Payload != "" {
		hr.Payload = []byte(a.RArgs.Payload)

	}

	// basic认证部分==========
	hr.Username, hr.Password = a.BasicAuth()

	// 合并
	hrall.ReqOpts = append(hrall.ReqOpts, hr)

	return hrall
}

const (
	maxResult    = 1000000
	maxIdleConn  = 500
	headerRegexp = `^([\w-]+):\s*(.+)`
	authRegexp   = `^(.+):([^\s].+)`
	heyUA        = "hey/0.0.2"
)

type ReqOpt struct {
	Url      *string
	Methord  *string
	Header   *http.Header
	Payload  []byte
	Username *string
	Password *string
	Req      *http.Request // 后面干掉。
	// Clt      *http.Client
}

type SingleTask struct {
	ReqOpts []*ReqOpt
	// Clt     *http.Client
}

type TaskFull struct {
	InitOnce  sync.Once
	Results   chan *result
	StopCh    chan struct{}
	Start     time.Duration
	Report    *report
	Outwriter io.Writer
	Output    string

	TArgs TaskArgs
	Task  SingleTask
	Clt   *http.Client
}

func NewTaskFull() *TaskFull {
	return &TaskFull{}
}

func (tf *TaskFull) NewClient(a *Args) *http.Client {
	tr := http.Transport{}
	certs := tls.Certificate{}
	if a.CArgs.Certfile != "" && a.CArgs.Keyfile != "" {
		fmt.Println("------111")
		certstmp, err := tls.LoadX509KeyPair(a.CArgs.Certfile, a.CArgs.Keyfile)
		if err != nil {
			fmt.Println(err)
		} else {
			certs = certstmp
		}
		ca, err := x509.ParseCertificate(certs.Certificate[0])
		if err != nil {
			fmt.Println(err)
		}
		pool := x509.NewCertPool()
		pool.AddCert(ca)

		tr = http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{certs},

				InsecureSkipVerify: true,
				ServerName:         a.RArgs.HostHeader,
			},
			MaxIdleConnsPerHost: utils.Min(a.TArgs.Conc, maxIdleConn),
			DisableCompression:  a.CArgs.DisableCompression,
			DisableKeepAlives:   a.CArgs.DisableKeepAlives,
			Proxy:               http.ProxyURL(utils.ParseURL(a.CArgs.ProxyAddr)),
		}
	} else {
		fmt.Println("------222")
		tr = http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         a.RArgs.HostHeader,
			},
			MaxIdleConnsPerHost: utils.Min(a.TArgs.Conc, maxIdleConn),
			DisableCompression:  a.CArgs.DisableCompression,
			DisableKeepAlives:   a.CArgs.DisableKeepAlives,
			Proxy:               http.ProxyURL(utils.ParseURL(a.CArgs.ProxyAddr)),
		}
	}

	if a.CArgs.H2 {
		http2.ConfigureTransport(&tr)
	} else {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}

	clt := &http.Client{
		Transport: &tr,
		Timeout:   time.Duration(a.TArgs.Timeout) * time.Second,
	}

	if a.CArgs.DisableRedirects {
		clt.CheckRedirect =
			func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
	}

	tf.Clt = clt
	return clt
}

func (tf *TaskFull) Writer() io.Writer {
	if tf.Outwriter == nil {
		return os.Stdout
	}
	return tf.Outwriter
}

func (tf *TaskFull) Init() {
	tf.InitOnce.Do(
		func() {
			tf.Results = make(chan *result, utils.Min(tf.TArgs.Conc*1000, maxResult))
			tf.StopCh = make(chan struct{}, tf.TArgs.Conc*1000)
		},
	)
}

func (tf *TaskFull) Run(clt *http.Client) {

	tf.Init()
	tf.Start = utils.Now()
	tf.Report = newReport(tf.Writer(), tf.Results, tf.Output, tf.TArgs.Nums)
	// Run the reporter first, it polls the result channel until it is closed.
	go func() {
		runReporter(tf.Report)
	}()
	tf.RunWorkers()
	tf.Finish()

}

func (tf *TaskFull) RunWorkers() {
	// Ignore the case where b.N % b.C != 0.
	var wg sync.WaitGroup
	wg.Add(tf.TArgs.Conc)
	for i := 0; i < tf.TArgs.Conc; i++ {
		go func() {
			tf.RunWorker(tf.TArgs.Nums / tf.TArgs.Conc) //注意此处去余了，也就是Ignore the case where b.N % b.C != 0
			wg.Done()
		}()
	}
	wg.Wait()
}

func (tf *TaskFull) RunWorker(n int) {
	var throttle <-chan time.Time
	if tf.TArgs.Qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(tf.TArgs.Qps)) * time.Microsecond) // 1e6/(b.QPS) 100w毫秒即1秒 / 1秒运行多少次= 一次运行的时间 即每次需要间隔多久才能达到这个qps
	}

	for i := 0; i < n; i++ {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-tf.StopCh:
			return
		default:
			if tf.TArgs.Qps > 0 {
				<-throttle //外层有N个runWorker的并发数，此函数是一个worker要访问多少次，如果没有sleep就一股脑发过去了
				//如果通过sleep变相控制了每秒访问的数量因此-n 1000 -c 100 -q 2 则是一秒访问100*2次 且 c * q < n ，否则n太小的话不到1s没意义，qps也不宜过大，超过本身性能极限，具体真实值查看  Requests/sec
			}
			tf.RequestRun()
		}
	}
}

func (tf *TaskFull) Finish() {
	close(tf.Results)
	total := utils.Now() - tf.Start
	// Wait until the reporter is done.
	<-tf.Report.done
	tf.Report.finalize(total)
}

func (tf *TaskFull) RequestRun() {
	for _, job := range tf.Task.ReqOpts {
		job.NewRequest()
		s := utils.Now()
		var size int64
		var code int
		var dnsStart, connStart, resStart, reqStart, delayStart time.Duration
		var dnsDuration, connDuration, resDuration, reqDuration, delayDuration time.Duration

		trace := &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) {
				dnsStart = utils.Now()
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				dnsDuration = utils.Now() - dnsStart
			},
			GetConn: func(h string) {
				connStart = utils.Now()
			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				if !connInfo.Reused {
					connDuration = utils.Now() - connStart
				}
				reqStart = utils.Now()
			},
			WroteRequest: func(w httptrace.WroteRequestInfo) {
				reqDuration = utils.Now() - reqStart
				delayStart = utils.Now()
			},
			GotFirstResponseByte: func() {
				delayDuration = utils.Now() - delayStart
				resStart = utils.Now()
			},
		}
		job.Req = job.Req.WithContext(httptrace.WithClientTrace(job.Req.Context(), trace))
		resp, err := tf.Clt.Do(job.Req)
		if err == nil {
			size = resp.ContentLength
			code = resp.StatusCode
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		} else {
			fmt.Println(err, "---RequestRun")
		}
		t := utils.Now()
		resDuration = t - resStart
		finish := t - s
		tf.Results <- &result{
			offset:        s,
			statusCode:    code,
			duration:      finish,
			err:           err,
			contentLength: size,
			connDuration:  connDuration,
			dnsDuration:   dnsDuration,
			reqDuration:   reqDuration,
			resDuration:   resDuration,
			delayDuration: delayDuration,
		}

		// resp, err = tf.Clt.Do(job.Req)
		// if err != nil {
		// 	fmt.Println(err, "task.clt.do")
		// }
		// bodybyte, err := ioutil.ReadAll(resp.Body)
		// fmt.Println(err, "taskgen-1--------")
		// fmt.Println(string(bodybyte))
	}
}

func (ro *ReqOpt) NewRequest() *http.Request {
	body := bytes.NewReader(ro.Payload)
	req, err := http.NewRequest(*ro.Methord, *ro.Url, body)
	if err != nil {
		utils.UsageAndExit(err.Error())
	}
	req.Header = *ro.Header

	req.ContentLength = int64(len(ro.Payload))
	if ro.Username != nil || ro.Password != nil {
		req.SetBasicAuth(*ro.Username, *ro.Password)
	}

	ro.Req = req
	return req
}

// func (t *SingleTask) NewRequest() *http.Request {
// 	req, _ := http.NewRequest("", "", nil)
// 	t.Req = req
// 	return req
// }

// func (ro *ReqOpt) GenRequest(t *SingleTask) *http.Request {
// 	t.Req.URL = utils.ParseURL(*ro.Url)
// 	t.Req.Methord = *ro.Methord
// 	t.Req.Header = *ro.Header
// 	if ro.Payload != nil {
// 		t.Req.Body.Read(ro.Payload)
// 	}

// 	t.Req.ContentLength = int64(len(ro.Payload))
// 	if ro.Username != nil || ro.Password != nil {
// 		t.Req.SetBasicAuth(*ro.Username, *ro.Password)
// 	}

// 	ro.Req = t.Req
// 	return t.Req
// }

type HeaderSlice []string

func (h *HeaderSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *HeaderSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}

package requester

import (
	"aloha/pkg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"time"
)

type result struct {
	err           error
	statusCode    int
	offset        time.Duration
	duration      time.Duration
	connDuration  time.Duration // connection setup(DNS lookup + Dial up) duration
	dnsDuration   time.Duration // dns lookup duration
	reqDuration   time.Duration // request "write" duration
	resDuration   time.Duration // response "read" duration
	delayDuration time.Duration // delay between response and request
	contentLength int64
}

type ReqJob struct {
	// Ropt    *taskargs.ReqOpt
	Request *http.Request
	Client  *http.Client
	results chan *result
}

// func (rj *ReqJob) GenClient(clt *http.Client) {
// 	rj.Client = clt
// }

// func (rj *ReqJob) GenRequest(req *http.Request) {
// 	rj.Request = req
// }

func (rj *ReqJob) Do() {
	var size int64
	var code int
	var dnsStart, connStart, resStart, reqStart, delayStart time.Duration
	var dnsDuration, connDuration, resDuration, reqDuration, delayDuration time.Duration
	req := rj.Request
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
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	s := utils.Now()

	resp, err := rj.Client.Do(req)
	if err == nil {
		size = resp.ContentLength
		code = resp.StatusCode
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
	t := utils.Now()
	resDuration = t - resStart
	finish := t - s
	rj.results <- &result{
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
}

// func userKill(rj *ReqJob) {
// 	// 处理用户终止ctrl-c，调用stop
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt)
// 	go func() {
// 		<-c
// 		rj.Stop()
// 	}()
// }

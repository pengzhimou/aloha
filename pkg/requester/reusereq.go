package requester

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

type Repeat struct {
	reader io.ReaderAt
	offset int64
}

// Read 重写读方法，使每次读取request.Body时能从指定位置读取
func (p *Repeat) Read(val []byte) (n int, err error) {
	n, err = p.reader.ReadAt(val, p.offset)
	atomic.AddInt64(&p.offset, int64(n))
	return
}

// Reset 重置偏移量
func (p *Repeat) Reset() {
	atomic.StoreInt64(&p.offset, 0)
}

func (p *Repeat) Close() error {
	// 因为req.Body实现了readcloser接口，所以要实现close方法
	// 但是repeat中的值有可能是只读的，所以这里只是尝试关闭一下。
	if p.reader != nil {
		if rc, ok := p.reader.(io.Closer); ok {
			return rc.Close()
		}
	}
	return nil
}

func doPost() {
	client := http.Client{}
	reader := strings.NewReader("hello")
	req, _ := http.NewRequest("POST", "http://www.abc.com", reader)
	req.Body = &Repeat{reader: reader, offset: 0}
	client.Do(req)
	// 将偏移量重置为0
	req.Body.(*Repeat).Reset()
	client.Do(req)
}

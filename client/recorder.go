package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type recorder struct {
	byteLen int64
	start   time.Time
	lapch   chan Lap
}

func newRecorder(s time.Time, maxConn int) *recorder {
	return &recorder{
		start: s,
		lapch: make(chan Lap, maxConn),
	}
}

func (r *recorder) Lap() <-chan Lap {
	return r.lapch
}

var src = rand.NewSource(0)

func (r *recorder) download(ctx context.Context, url string, size int) error {
	url = fmt.Sprintf("%s?size=%d", url, size)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx) // Attach context to request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	//　サーバーからダウンロードを少しづつ行い、そのスループットのベンチマークをとる
	proxy := r.newMeasureProxy(ctx, resp.Body)
	if _, err := io.Copy(io.Discard, proxy); err != nil {
		return err
	}

	return nil
}

// Proxy for measure benchmark
type measureProxy struct {
	io.Reader
	*recorder // embeded
}

func (r *recorder) newMeasureProxy(ctx context.Context, reader io.Reader) io.Reader {
	rp := &measureProxy{
		Reader:   reader,
		recorder: r,
	}
	go rp.Watch(ctx, r.lapch)
	return rp
}

func (m *measureProxy) Watch(ctx context.Context, send chan<- Lap) {
	t := time.NewTicker(150 * time.Millisecond)
	// 150ミリ秒ごとにスループットをチェック
	for {
		select {
		case <-t.C:
			byteLen := atomic.LoadInt64(&m.byteLen)
			delta := time.Now().Sub(m.start).Seconds()
			send <- newLap(byteLen, delta)
		case <-ctx.Done():
			return
		}
	}
}

// Read server downloads and add byteLen
func (m *measureProxy) Read(p []byte) (n int, err error) {
	n, err = m.Reader.Read(p)
	atomic.AddInt64(&m.byteLen, int64(n))
	return
}

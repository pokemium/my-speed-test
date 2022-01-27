package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

var DownloadTimeout = 15 * time.Second

const downloadURL = api + "/download"

func DownloadTest(ctx context.Context, cb IntervalCallback) error {
	// タイムアウトとエラー時にコンテキストキャンセルを実行できるようにする
	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	r := newRecorder(time.Now(), maxConnections)

	// ラップのポーリング
	go func() {
		for {
			select {
			case lap := <-r.Lap():
				fmt.Println(&lap)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 同時最大接続数の制限のためのセマフォ
	semaphore := make(chan struct{}, maxConnections)

loop:
	for _, size := range payloadSizes {
		// Allocate new connection.
		select {
		case <-ctx.Done():
			break loop

		// If semaphore is full, pending connection.
		case semaphore <- struct{}{}:
			time.Sleep(250 * time.Millisecond)
		}

		eg.Go(func() error {
			defer func() { <-semaphore }() // Decrease semaphore
			err := r.download(ctx, downloadURL, size)
			if err != nil {
				return err
			}
			return nil
		})
	}

	select {
	case <-ctx.Done():
	case semaphore <- struct{}{}:
		cancel()
	}

	return eg.Wait()
}

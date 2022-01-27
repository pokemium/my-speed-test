package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

const maxSize = 1024

func downloadHandler() http.HandlerFunc {
	src := rand.NewSource(0)

	// ダウンロードエンドポイントではランダム文字列を返す
	return func(w http.ResponseWriter, r *http.Request) {
		queries := r.URL.Query()
		size := queries.Get("size")
		max, err := strconv.Atoi(size)
		if err != nil {
			max = maxSize
		}

		read := rand.New(src)
		_, err = io.CopyN(w, read, int64(max))
		if err != nil {
			log.Printf("failed to write random data: %s", err)
			return
		}
	}
}

func uploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentLength := r.ContentLength
		if contentLength > maxSize {
			contentLength = maxSize
		}

		_, err := io.CopyN(io.Discard, r.Body, contentLength)
		if err != nil {
			log.Printf("failed to write body: %s", err)
			return
		}
	}
}

func main() {

}

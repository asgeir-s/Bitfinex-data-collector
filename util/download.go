package util

import (
	"io"
	"net/http"
	"os"
	"time"
	"math/big"
	""
)

// DownloadIfModifiedSince file from url to fileName
func DownloadIfModifiedSince(url string, fileName string, since time.Time) (int64, error) {
	// download file
	out, err := os.Create(fileName)

	if err != nil {
		return 0, err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	lastMod, err := time.Parse("Mon, 02 Jan 2006 15:04:00 MST", resp.Header.Get("Last-Modified"))
	if err != nil {
		return 0, err
	}

	test1 := new(big.Float)
	test1.SetString("1")

	test2 := new(big.Float)
	test2.SetString("1.000009999")

	test3 := new(big.Float)
	test3.SetString("5")

	num, int, err := big.ParseFloat("1", 10, 100, big.ToNearestEven)
	println("num", num)
	println("int", int)
	println("err", err)

	if lastMod.After(since) {
		n, err := io.Copy(out, resp.Body)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	return 0, nil
}



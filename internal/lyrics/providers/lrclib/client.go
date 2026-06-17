package lrclib

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	errs "lrcsnc/internal/lyrics/errors"
	"lrcsnc/internal/pkg/global"
	"lrcsnc/internal/pkg/log"
)

var userAgent string = fmt.Sprintf("lrcsnc %v (https://github.com/Endg4meZer0/lrcsnc)", global.Version)

var httpClient = http.Client{
	Timeout: 15 * time.Second,
}

func getLyrics(title string, artist string, album string, duration float64) ([]byte, error) {
	params := url.Values{}
	params.Set("track_name", title)
	params.Set("artist_name", artist)
	params.Set("album_name", album)
	params.Set("duration", fmt.Sprintf("%v", int(duration)))
	urlReqPath := "https://lrclib.net/api/get?" + params.Encode()
	_, err := url.Parse(urlReqPath)
	if err != nil {
		log.Fatal("lyrics/providers/lrclib/client", fmt.Sprintf("Failed to parse string (%v) to URL:\n%v", urlReqPath, err))
	}
	req, err := http.NewRequest(http.MethodGet, urlReqPath, nil)
	if err != nil {
		log.Fatal("lyrics/providers/lrclib/client", fmt.Sprintf("Failed to make http.Request: %v", err))
	}
	req.Header.Add("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if os.IsTimeout(err) {
		return nil, errs.ServerTimeout
	}
	if err != nil || resp.StatusCode != 200 {
		return nil, errs.ServerError
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errs.BodyReadFail
	}
	return body, nil
}

func searchLyrics(title string, artist string) ([]byte, error) {
	params := url.Values{}
	params.Set("track_name", title)
	params.Set("artist_name", artist)
	urlReqPath := "https://lrclib.net/api/search?" + params.Encode()
	_, err := url.Parse(urlReqPath)
	if err != nil {
		log.Fatal("lyrics/providers/lrclib/client", fmt.Sprintf("Failed to parse string (%v) to URL; please, report this issue to GitHub. More:\n%v", urlReqPath, err))
	}

	resp, err := httpClient.Get(urlReqPath)
	if os.IsTimeout(err) {
		return nil, errs.ServerTimeout
	}
	if err != nil || resp.StatusCode != 200 {
		return nil, errs.ServerError
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errs.BodyReadFail
	}
	return body, nil
}

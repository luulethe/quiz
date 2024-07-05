package cache_wrapper

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func compressString(str string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(str)); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}

func decompressString(str string) (string, error) {
	r, err := gzip.NewReader(bytes.NewReader([]byte(str)))
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

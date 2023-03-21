package xcrypto

import (
	"bytes"
	"compress/gzip"
)

func Gzip(str string) (string, error) {
	var bf bytes.Buffer
	w := gzip.NewWriter(&bf)
	if _, err := w.Write([]byte(str)); err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return bf.String(), nil
}

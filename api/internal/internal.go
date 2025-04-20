package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status code %d, '%s'\n", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func WriteFile(name string, content []byte) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(content)
	return err
}

func ReadOptions(options ...string) ([]byte, error) {
	var e error

	for _, o := range options {
		if content, err := os.ReadFile(o); err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				return []byte{}, err
			} else if e == nil {
				e = err
			}
		} else {
			return content, nil
		}
	}

	return []byte{}, errors.New(strings.ReplaceAll(e.Error(), options[0], "'"+strings.Join(options, "' or '")+"'"))
}

package mime

import (
	mimer "mime"
	"path/filepath"
)

func DetectContentType(sample []byte) (string, error) {
	m, e := detectContentType(sample)
	if e != nil {
		return "", e
	}
	return m, nil
}

func DetectContentTypeFromPath(path string) (string, error) {
	ext := filepath.Ext(path)
	m := mimer.TypeByExtension(ext)
	if m == "" {
		return detectContentTypeFromPath(path)
	}
	return m, nil
}

package keyval

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"
)

var (
	ErrNotFound = errors.New("not found")
)

type ValueInfo struct {
	Size int64
	Hash []byte
}

type KeyValStoreFactory func(options interface{}) (KeyValStore, error)

type KeyValStore interface {
	Set(key []byte, reader io.Reader) error
	SetBytes(key []byte, bytes []byte) error
	Has(bs []byte) bool
	Remove(key []byte) bool
	Get(key []byte) (io.ReadCloser, error)
	GetBytes(key []byte) ([]byte, error)
}

type ResourceStore interface {
	KeyValStore
	OriginalUrl(key []byte) string
	State(key []byte) (ValueInfo, error)
}

var _store map[string]KeyValStoreFactory

func init() {
	_store = make(map[string]KeyValStoreFactory)
}

func Register(name string, fn KeyValStoreFactory) {
	_store[name] = fn
}

func Store(name string, options interface{}) (KeyValStore, error) {
	if s, ok := _store[name]; ok {
		return s(options)
	}
	return nil, fmt.Errorf("store: '%s' not found", name)
}

func GetOptions(options interface{}, out interface{}) error {

	var err error
	switch t := options.(type) {
	case []byte:
		err = json.Unmarshal(t, out)
	case string:
		err = json.Unmarshal([]byte(t), out)
	case map[string]interface{}:
		err = mapstructure.Decode(t, out)
	}

	return err
}

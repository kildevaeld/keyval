package keyval

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mitchellh/mapstructure"
)

var (
	ErrNotFound = errors.New("not found")
	ErrStopIter = errors.New("stop iterater")
)

/*type ValueInfo struct {
	Size int64
	Hash []byte
}*/

type KeyValStoreFactory func(options interface{}) (KeyValStore, error)

type KeyValStore interface {
	Set(key []byte, reader io.Reader) error
	SetBytes(key []byte, bytes []byte) error
	Has(bs []byte) bool
	Remove(key []byte) bool
	Get(key []byte) (io.ReadCloser, error)
	GetBytes(key []byte) ([]byte, error)
}

type Stat interface {
	Size() int64
	Mtime() time.Time
	Ctime() time.Time
	Hash() []byte
	IsDir() bool
}

type KeyValMetaStore interface {
	Stat([]byte) (Stat, error)
	List(prefix []byte, fn func(key []byte, meta Stat) error) error
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

type stat_impl struct {
	size  int64
	hash  []byte
	ctime time.Time
	mtime time.Time
	isDir bool
}

func (s *stat_impl) Size() int64 {
	return s.size
}
func (s *stat_impl) Mtime() time.Time {
	return s.mtime
}
func (s *stat_impl) Ctime() time.Time {
	return s.ctime
}
func (s *stat_impl) Hash() []byte {
	return s.hash
}
func (s *stat_impl) IsDir() bool {
	return s.isDir
}

func NewState(s int64, h []byte, c time.Time, m time.Time) Stat {
	return &stat_impl{
		s, h, c, m, false,
	}
}

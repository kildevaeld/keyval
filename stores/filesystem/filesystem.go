package filesystem

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	system "github.com/kildevaeld/go-system"
	"github.com/kildevaeld/keyval"
)

func hasParent(path string) bool {
	parent := filepath.Dir(path)
	return parent != "." && parent != "/"
}

type FileSystemOptions struct {
	Path     string `json:"path"`
	HashKeys string `json:"hash_keys,omitempty" mapstructure:"hash_keys"`
}

type filesystem struct {
	path     string
	hashKeys string
}

func (f *filesystem) mkDir(key string) error {
	if !hasParent(key) {
		return nil
	}

	parent := filepath.Dir(key)

	if _, err := os.Stat(parent); err != nil {
		logrus.Debugf("Create directory: %s", parent)
		return os.MkdirAll(parent, 0777)
	}
	return nil

}

func (f *filesystem) Set(key []byte, reader io.Reader) error {

	str := f.key(key)

	if err := f.mkDir(str); err != nil {
		return err
	}

	file, err := os.Create(str)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)

	return err
}

func (f *filesystem) SetBytes(key []byte, bs []byte) error {
	return f.Set(key, bytes.NewReader(bs))
}

func (f *filesystem) Has(bs []byte) bool {
	_, err := os.Stat(f.key(bs))
	return err != nil
}

func (f *filesystem) Remove(key []byte) bool {
	return os.Remove(f.key(key)) != nil
}

func (f *filesystem) Get(key []byte) (io.ReadCloser, error) {
	reader, err := os.Open(f.key(key))
	if err != nil {
		return nil, keyval.ErrNotFound
	}
	return reader, nil
}

func (f *filesystem) GetBytes(key []byte) ([]byte, error) {
	file, err := f.Get(key)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func (f *filesystem) key(key []byte) string {
	if f.hashKeys != "" {
		var hash hash.Hash
		switch f.hashKeys {
		case "sha256":
			hash = sha256.New()
		case "sha512":
			hash = sha512.New()
		default:
			panic(fmt.Sprintf("invalid algorithm: %s", f.hashKeys))
		}

		return fmt.Sprintf("%s/%x", f.path, hash.Sum(key))

	}
	return filepath.Join(f.path, string(key))
}

func (f *filesystem) init() (*filesystem, error) {

	if !filepath.IsAbs(f.path) {
		path, err := filepath.Abs(f.path)
		if err != nil {
			return nil, err
		}
		f.path = path
	}

	if info, err := os.Stat(f.path); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("path '%s' already exists, and is not a directory", f.path)
		}
	}

	if err := os.MkdirAll(f.path, 0770); err != nil {
		return nil, err
	}

	return f, nil
}

func init() {
	keyval.Register("filesystem", func(options interface{}) (keyval.KeyValStore, error) {
		if options == nil {
			return nil, fmt.Errorf("FileSystem store needs a path parameter")
		}

		var (
			o  FileSystemOptions
			ok bool
		)

		if o, ok = options.(FileSystemOptions); !ok {
			if err := keyval.GetOptions(options, &o); err != nil {
				return nil, err
			}
		}

		if o.Path == "" || o.Path == "." || o.Path == "/" {
			return nil, errors.New("path cannot not be empty")
		}

		o.Path = system.Environ(os.Environ()).Expand(o.Path)

		f := &filesystem{
			path:     o.Path,
			hashKeys: o.HashKeys,
		}

		return f.init()
	})
}

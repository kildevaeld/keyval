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

	"go.uber.org/zap"

	system "github.com/kildevaeld/go-system"
	"github.com/kildevaeld/keyval"
	"github.com/vmihailenco/msgpack"
)

var (
	metaKeyName = "__meta"
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
	info     map[string]*Info
}

func (f *filesystem) mkDir(key string) error {
	if !hasParent(key) {
		return nil
	}

	parent := filepath.Dir(key)

	if _, err := os.Stat(parent); err != nil {
		zap.L().Sugar().Debugf("Create directory: %s", parent)
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

	if err != nil {
		s, e := os.Stat(str)
		if e != nil {
			return e
		}
		return f.add(string(key), &Info{
			size:  s.Size(),
			ctime: s.ModTime(),
			mtime: s.ModTime(),
			hash:  nil,
		})
	}

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

func (f *filesystem) Stat(key []byte) (keyval.Stat, error) {

	info, err := os.Stat(f.key(key))
	if err != nil {
		if err != os.ErrNotExist {
			return nil, keyval.ErrNotFound
		}
		return nil, err
	} else if info.IsDir() {
		return &Info{
			size:  info.Size(),
			mtime: info.ModTime(),
			isDir: true,
		}, nil
	}

	return f.info[string(key)], nil
}
/*
func (f *filesystem) List(prefix []byte, fn func(key []byte, meta keyval.Stat) error) error {
	path := f.key(prefix)
	files, err := filepath.Glob(string(path))
	if err != nil {
		return err
	}

	for _, k := range files {
		//f,e := os.Stat(filepath.Join(string(path), k))
	}

	return nil
}*/

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

	f.load()

	return f, nil
}

func (f *filesystem) load() error {
	bs, err := ioutil.ReadFile(filepath.Join(f.path, metaKeyName))
	if err != nil {
		return nil
	}

	if err := msgpack.Unmarshal(bs, &f.info); err != nil {
		return nil
	}

	return nil
}

func (f *filesystem) save() error {
	var (
		bs  []byte
		err error
	)
	if bs, err = msgpack.Marshal(f.info); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(f.path, metaKeyName), bs, 0666)

}

func (f *filesystem) add(p string, v *Info) error {
	f.info[p] = v
	return f.save()
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
			info:     make(map[string]*Info),
		}

		return f.init()
	})
}

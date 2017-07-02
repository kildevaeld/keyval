package memory

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/kildevaeld/keyval"
)

type memory struct {
	mem map[string][]byte
}

func (m *memory) Set(key []byte, reader io.Reader) error {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	return m.SetBytes(key, bs)
}

func (m *memory) SetBytes(key []byte, bytes []byte) error {
	m.mem[string(key)] = bytes
	return nil
}
func (m *memory) Has(bs []byte) bool {
	_, ok := m.mem[string(bs)]
	return ok
}
func (m *memory) Remove(key []byte) bool {
	_, ok := m.mem[string(key)]
	delete(m.mem, string(key))
	return ok
}
func (m *memory) Get(key []byte) (io.ReadCloser, error) {
	bs, err := m.GetBytes(key)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bs)), nil
}

func (m *memory) GetBytes(key []byte) ([]byte, error) {
	bs, ok := m.mem[string(key)]
	if !ok {
		return nil, keyval.ErrNotFound
	}
	return bs, nil
}

func init() {
	keyval.Register("memory", func(options interface{}) (keyval.KeyValStore, error) {
		return &memory{
			mem: make(map[string][]byte),
		}, nil
	})
}

package filesystem

import (
	"time"

	"github.com/kildevaeld/dict"
	"github.com/kildevaeld/keyval"
	"github.com/vmihailenco/msgpack"
)

type Info struct {
	size  int64
	hash  []byte
	ctime time.Time
	mtime time.Time
	isDir bool
}

func (s *Info) Size() int64 {
	return s.size
}
func (s *Info) Mtime() time.Time {
	return s.mtime
}
func (s *Info) Ctime() time.Time {
	return s.ctime
}
func (s *Info) Hash() []byte {
	return s.hash
}
func (s *Info) IsDir() bool {
	return s.isDir
}

func (s *Info) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(dict.Map{
		"size":  s.size,
		"hash":  s.hash,
		"ctime": s.ctime,
		"mtime": s.mtime,
	})
}

func (s *Info) UnmarshalMsgpack(bs []byte) error {
	var m dict.Map
	if err := msgpack.Unmarshal(bs, &m); err != nil {
		return err
	}
	s.ctime = m.Get("ctime").(time.Time)
	s.mtime = m.Get("mtime").(time.Time)
	s.size = m.Get("size").(int64)
	s.hash = m.Get("hash").([]byte)
	return nil
}

func NewState(s int64, h []byte, c time.Time, m time.Time, d bool) keyval.Stat {
	return &Info{
		s, h, c, m, d,
	}
}

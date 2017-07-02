package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aarzilli/golua/lua"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kildevaeld/bproxy/mime"
	"github.com/kildevaeld/keyval"
	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse"
	luam "github.com/kildevaeld/valse/middlewares/lua"
	"github.com/stevedonovan/luar"
)

var FileField = "file"

type ServerOptions struct {
	ScriptPath string
	WorkQueue  int
}

type HttpServer struct {
	v  *valse.Server
	kv keyval.KeyValStore
}

func (s *HttpServer) Listen(addr string) error {
	return s.v.Listen(addr)
}

func (s *HttpServer) initLua() *lua.State {
	L := luar.Init()
	L.OpenLibs()

	luar.GoToLua(L, luar.Map{
		"jwt": luar.Map{
			"decode": func(tokenString, key string) (luar.Map, error) {
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(key), nil
				})
				if err != nil {
					return nil, err
				}

				return luar.Map{
					"algo":   token.Method.Alg(),
					"valid":  token.Valid,
					"claims": token.Claims.(jwt.MapClaims),
				}, nil
			},
		},
	})
	L.SetGlobal("kv")

	return L
}

func (s *HttpServer) init(options ServerOptions) error {

	//s.v.Use(middlewares.NewWithNameAndLogrus("http", logrus.WithField("", "")))
	if options.ScriptPath != "" {
		s.v.Use(luam.Lua(luam.LuaOptions{
			Path:       options.ScriptPath,
			LuaFactory: s.initLua,
			WorkQueue:  options.WorkQueue,
		}))
	}

	s.v.Get("/*path", s.handleGet)
	s.v.Head("/*path", s.handleCheck)
	s.v.Post("/*path", s.handleSet)

	return nil
}

func (s *HttpServer) handleCheck(ctx *valse.Context) error {

	name := ctx.UserValue("path").(string)
	if name == "/" {
		return strong.NewHTTPError(strong.StatusBadRequest)
	}

	if !s.kv.Has([]byte(name)) {
		ctx.SetStatusCode(strong.StatusOK)
	} else {
		ctx.SetStatusCode(strong.StatusNotFound)
	}

	if i, ok := s.kv.(keyval.ResourceStore); ok {
		stat, err := i.State([]byte(name))
		if err != nil {
			return err
		}
		ctx.Response.Header.SetContentLength(int(stat.Size))
	}

	return nil
}

func (s *HttpServer) handleSet(ctx *valse.Context) error {

	name := ctx.UserValue("path").(string)
	if name == "/" {
		return strong.NewHTTPError(strong.StatusBadRequest)
	}
	var reader io.ReadCloser
	if bytes.HasPrefix(ctx.Request.Header.ContentType(), []byte("multipart/form-data")) {
		file, err := ctx.FormFile(FileField)
		if err != nil {
			return err
		}
		if reader, err = file.Open(); err != nil {
			return err
		}

	} else {
		reader = ioutil.NopCloser(bytes.NewReader(ctx.PostBody()))
	}

	defer reader.Close()

	err := s.kv.Set([]byte(name), reader)

	return err
}

func (s *HttpServer) handleGet(ctx *valse.Context) error {

	name := ctx.UserValue("path").(string)
	if name == "/" {
		return strong.NewHTTPError(strong.StatusBadRequest)
	}

	if i, ok := s.kv.(keyval.ResourceStore); ok {
		stat, err := i.State([]byte(name))
		if err != nil {
			return err
		}
		ctx.Response.Header.SetContentLength(int(stat.Size))
		ctx.Response.Header.Set("ETag", fmt.Sprintf("\"%x\"", stat.Hash))
	}

	file, err := s.kv.Get([]byte(name[1:]))
	if err != nil {
		return err
	}
	var (
		bs [64]byte
		i  int
		e  error
		m  string
	)

	if i, e = file.Read(bs[:]); e != nil {
		return e
	}

	if m, e = mime.DetectContentType(bs[:]); e != nil {
		return e
	}
	ctx.Response.Header.Set(strong.HeaderContentType, m)
	ctx.Write(bs[:i])

	if i < 64 {
		return nil
	}
	_, e = io.Copy(ctx, file)

	return e
}

func NewServer(kv keyval.KeyValStore, options ServerOptions) (*HttpServer, error) {

	s := &HttpServer{v: valse.New(), kv: kv}

	err := s.init(options)

	if err != nil {
		return nil, err
	}

	return s, nil
}

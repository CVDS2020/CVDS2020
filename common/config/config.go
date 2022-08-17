package config

import (
	"encoding/json"
	"github.com/CVDS2020/CVDS2020/common/errors"
	"github.com/CVDS2020/CVDS2020/common/id"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
)

const MaxConfigFileSize = 32 * 1024 * 1024

type handler func(c interface{}) (nc interface{}, err error)

type PreHandlerConfig interface {
	PreHandle() PreHandlerConfig
}

type PostHandlerConfig interface {
	PostHandle() (PostHandlerConfig, error)
}

func preHandler(c interface{}) (nc interface{}, err error) {
	if pc, ok := c.(PreHandlerConfig); ok {
		return pc.PreHandle(), nil
	}
	return c, nil
}

func postHandler(c interface{}) (nc interface{}, err error) {
	if pc, ok := c.(PostHandlerConfig); ok {
		return pc.PostHandle()
	}
	return c, nil
}

func handle(handler handler, c interface{}, cs map[interface{}]struct{}) (nc interface{}, err error) {
	if _, repeat := cs[c]; repeat {
		return c, nil
	}

	cs[c] = struct{}{}
	if nc, err = handler(c); err != nil {
		return nil, err
	} else if nc != c {
		delete(cs, c)
		cs[c] = struct{}{}
	}

	v := reflect.ValueOf(nc)
	if v.Kind() == reflect.Ptr {
		// value is pointer, try get element value
		if v.IsNil() {
			// value is nil pointer
			t := v.Type().Elem()
			if t.Kind() == reflect.Struct {
				// value is nil struct pointer, walk fields and handle field, if handle field
				// return not nil, create this struct and set field
				nf := t.NumField()
				for i := 0; i < nf; i++ {
					if f := t.Field(i); f.IsExported() {
						var zf reflect.Value
						if f.Type.Kind() == reflect.Ptr {
							zf = reflect.Zero(f.Type)
						} else {
							zf = reflect.NewAt(f.Type, nil)
						}
						zfi := zf.Interface()
						if nzf, err := handle(handler, zfi, cs); err != nil {
							return nil, err
						} else if nzf != zfi {
							if v.IsNil() {
								v = reflect.New(t)
								nc = v.Interface()
							}
							if f.Type.Kind() == reflect.Ptr {
								v.Elem().Field(i).Set(reflect.ValueOf(nzf))
							} else {
								v.Elem().Field(i).Set(reflect.ValueOf(nzf).Elem())
							}
						}
					}
				}
			}
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		nf := v.NumField()
		// walk fields and handle filed
		for i := 0; i < nf; i++ {
			f := v.Field(i)
			if f.Kind() == reflect.Ptr {
				// field is pointer
				if f.CanInterface() {
					fi := f.Interface()
					if nfi, err := handle(handler, fi, cs); err != nil {
						return nil, err
					} else if nfi != fi {
						f.Set(reflect.ValueOf(nfi))
					}
				}
				continue
			}
			// field not pointer
			if f.CanAddr() {
				fp := f.Addr()
				if fp.CanInterface() {
					fpi := fp.Interface()
					if nfp, err := handle(handler, fpi, cs); err != nil {
						return nil, err
					} else if nfp != fpi {
						fp.Set(reflect.ValueOf(nfp))
					}
				}
			}
		}

	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			mv := iter.Value().Interface()
			if nmv, err := handle(handler, mv, cs); err != nil {
				return nil, err
			} else if nmv != mv {
				v.SetMapIndex(iter.Key(), reflect.ValueOf(nmv))
			}
		}

	case reflect.Slice, reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			lv := v.Index(i)
			lvi := lv.Interface()
			if nlv, err := handle(handler, lvi, cs); err != nil {
				return nil, err
			} else if nlv != lv {
				lv.Set(reflect.ValueOf(nlv))
			}
		}
	}

	return
}

type Type struct {
	Id          TypeId
	Name        string
	Suffixes    []string
	Unmarshaler func([]byte, interface{}) error
}

type TypeId uint64

var typeIdCxt uint64

func GenTypeId() TypeId {
	return TypeId(id.Uint64Id(&typeIdCxt))
}

var (
	TypeUnknown = Type{Id: GenTypeId(), Name: "unknown"}
	TypeYaml    = Type{Id: GenTypeId(), Name: "yaml", Suffixes: []string{"yaml", "yml"}, Unmarshaler: yaml.Unmarshal}
	TypeJson    = Type{Id: GenTypeId(), Name: "json", Suffixes: []string{"json"}, Unmarshaler: json.Unmarshal}
)

var types = []*Type{&TypeYaml, &TypeJson}

func ProbeType(path string) Type {
	for _, tpy := range types {
		for _, suffix := range tpy.Suffixes {
			if strings.HasSuffix(path, suffix) {
				return *tpy
			}
		}
	}
	return TypeUnknown
}

type Parser struct {
	configs []struct {
		path string
		data []byte
		typ  Type
	}
}

func Exist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.Mode().IsRegular()
}

func (p *Parser) SetConfigFile(path string, typ *Type) {
	var t Type
	if typ == nil {
		t = ProbeType(path)
	} else {
		t = *typ
	}
	p.configs = []struct {
		path string
		data []byte
		typ  Type
	}{{path: path, typ: t}}
}

func (p *Parser) AddConfigFile(path string, typ *Type) {
	var t Type
	if typ == nil {
		t = ProbeType(path)
	} else {
		t = *typ
	}
	p.configs = append(p.configs, struct {
		path string
		data []byte
		typ  Type
	}{path: path, typ: t})
}

func (p *Parser) Unmarshal(c interface{}) error {
	// invoke config pre handle
	handle(preHandler, c, make(map[interface{}]struct{}))
	for i, config := range p.configs {
		info, err := os.Stat(config.path)
		if err != nil {
			return err
		}
		if info.Size() > MaxConfigFileSize {
			return errors.New("config file size too large")
		}
		data, err := os.ReadFile(config.path)
		if err != nil {
			return err
		}
		p.configs[i].data = data
	}

out:
	// parse config
	for _, config := range p.configs {
		if config.typ.Id == TypeUnknown.Id {
			var es error
			for _, tpy := range types {
				if tpy.Unmarshaler != nil {
					if err := config.typ.Unmarshaler(config.data, c); err != nil {
						// retry next unmarshaler parse config
						errors.Append(es, err)
						continue
					}
					// parse config success
					continue out
				}
			}
			// all unmarshaler parse config failed
			return es
		} else {
			if err := config.typ.Unmarshaler(config.data, c); err != nil {
				return err
			}
		}
	}

	// invoke config post handle
	_, err := handle(postHandler, c, make(map[interface{}]struct{}))
	return err
}

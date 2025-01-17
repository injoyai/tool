package config

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"path/filepath"
)

const (
	STRING = "string"
	INT    = "int"
	FLOAT  = "float"
	BOOL   = "bool"
	SELECT = "select"
	MAP    = "map"
	OBJECT = "object"
	ARRAY  = "array"
)

func New(filename string, natures Natures) *Config {
	oss.NewDir(filepath.Dir(filename))
	m := cfg.WithFile(filename).(*conv.Map)
	return &Config{
		Filename: filename,
		Natures:  natures,
		m:        m,
	}
}

type Config struct {
	Width    int               //宽度,可选
	Height   int               //高度,可选
	Filename string            //本地文件路径,必须
	Natures  []*Nature         //格式,必须
	OnSaved  func(m *conv.Map) //保存事件,可选
	m        *conv.Map         //缓存数据
}

func (this *Config) SetWidthHeight(width, height int) *Config {
	this.Width = width
	this.Height = height
	return this
}

func (this *Config) SetOnSaved(onSaved func(m *conv.Map)) *Config {
	this.OnSaved = onSaved
	return this
}

func (this *Config) Get() []*Nature {
	if this.m == nil {
		this.m = cfg.WithFile(this.Filename).(*conv.Map)
	}
	return initNature(this.Natures, this.m)
}

func (this *Config) Save(m g.Map) error {
	if this.m == nil {
		this.m = cfg.WithFile(this.Filename).(*conv.Map)
	}
	for k, v := range m {
		this.m.Set(k, v)
	}
	err := oss.New(this.Filename, this.m.String())
	if err != nil {
		return err
	}
	if this.OnSaved != nil {
		this.OnSaved(this.m)
	}
	return nil
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"` //这个填写无效,会赋值为配置文件的值
}

type Natures []*Nature

func initNature(natures []*Nature, m *conv.Map) []*Nature {
	for i := range natures {
		switch natures[i].Type {
		case STRING:
			natures[i].Value = m.GetString(natures[i].Key)
		case INT:
			natures[i].Value = m.GetInt(natures[i].Key)
		case FLOAT:
			natures[i].Value = m.GetFloat64(natures[i].Key)
		case BOOL:
			natures[i].Value = m.GetBool(natures[i].Key)
		case SELECT:
			var ls []*Nature
			conv.Unmarshal(natures[i].Value, &ls)
			natures[i].Value = ls
		case MAP:
			object := Natures{}
			for k, v := range m.GetGMap(natures[i].Key) {
				object = append(object, &Nature{
					Name:  k,
					Key:   k,
					Value: v,
				})
			}
			natures[i].Value = object
		case OBJECT:
			if natures[i].Value == nil {
				natures[i].Value = []*Nature{}
			}
			var ls []*Nature
			conv.Unmarshal(natures[i].Value, &ls)
			for k, v := range m.GetGMap(natures[i].Key) {
				for j := range ls {
					if ls[j].Key == k {
						ls[j].Value = v
						continue
					}
				}
			}
			natures[i].Value = ls
		default:
			natures[i].Value = m.GetString(natures[i].Key)
		}
	}
	return natures
}

func (this Natures) Map() g.Map {
	m := g.Map{}
	for _, v := range this {
		m[v.Key] = v.Value
	}
	return m
}

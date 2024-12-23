package config

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"path/filepath"
)

func New(filename string, natures Natures) *Config {
	oss.NewDir(filepath.Dir(filename))
	m := cfg.WithFile(filename).(*conv.Map)
	return &Config{
		Filename: filename,
		Natures:  initNature(natures, m),
		m:        m,
	}
}

type Config struct {
	Filename string
	Natures  []*Nature
	m        *conv.Map
}

func (this *Config) Get() []*Nature {
	return this.Natures
	return initNature(this.Natures, this.m)
}

func (this *Config) Save(m g.Map) error {
	for k, v := range m {
		this.m.Set(k, v)
	}
	logs.Debug(this.Filename)
	logs.Debug(this.m.String())
	return oss.New(this.Filename, this.m.String())
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type Natures []*Nature

func initNature(natures []*Nature, m *conv.Map) []*Nature {
	for i := range natures {
		switch natures[i].Type {
		case "bool":
			natures[i].Value = m.GetBool(natures[i].Key)
		case "object":
			object := Natures{}
			for k, v := range m.GetGMap(natures[i].Key) {
				object = append(object, &Nature{
					Name:  k,
					Key:   k,
					Value: v,
				})
			}
			natures[i].Value = object
		case "object2":
			if natures[i].Value == nil {
				natures[i].Value = []*Nature{}
			}
			ls := natures[i].Value.([]Nature)
			for k, v := range m.GetGMap(natures[i].Key) {
				for j := range ls {
					if ls[j].Key == k {
						ls[j].Value = v
						continue
					}
				}
			}
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

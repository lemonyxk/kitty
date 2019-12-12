/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-22 21:31
**/

package plugin

import (
	"plugin"

	"github.com/Lemo-yxk/lemo/exception"
)

type Func func(v ...interface{}) (interface{}, exception.ErrorFunc)

type Plugin struct {
	p *plugin.Plugin
}

type Script struct {
	f plugin.Symbol
}

func Open(file string) *Plugin {
	p, err := plugin.Open(file)
	if err != nil {
		return &Plugin{}
	}
	return &Plugin{p: p}
}

func (p *Plugin) Lookup(symName string) *Script {
	if p.p == nil {
		return &Script{}
	}

	f, err := p.p.Lookup(symName)
	if err != nil {
		return &Script{}
	}

	return &Script{f: f}
}

func (s *Script) Run(v ...interface{}) (interface{}, exception.ErrorFunc) {
	if s.f == nil {
		return nil, exception.New("lookup error, please check the func name")
	}

	if f, ok := s.f.(func(v ...interface{}) (interface{}, exception.ErrorFunc)); ok {
		return f(v...)
	}

	return nil, exception.New("assert error, please check the func type")
}

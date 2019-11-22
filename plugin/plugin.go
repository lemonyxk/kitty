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
		return nil
	}

	f, err := p.p.Lookup(symName)
	if err != nil {
		return nil
	}

	return &Script{f: f}
}

func (s *Script) Run() func() *exception.Error {
	if s.f == nil {
		return nil
	}
	return s.f.(func(v ...interface{}) func() *exception.Error)()
}

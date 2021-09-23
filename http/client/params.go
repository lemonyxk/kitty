/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-21 17:36
**/

package client

import (
	"context"
	"net/http"
)

type params struct {
	info *info
	err  error
	req  *http.Request

	cancel context.CancelFunc
}

func (p *params) Send() *Req {
	if p.err != nil {
		return &Req{err: p.err}
	}
	return send(p.info, p.req, p.cancel)
}

func (p *params) Abort() {
	p.cancel()
}

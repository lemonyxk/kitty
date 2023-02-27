/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-21 17:36
**/

package client

import (
	"context"
	"net/http"
)

type Sender struct {
	info *Request
	err  error
	req  *http.Request

	cancel context.CancelFunc
}

func (p *Sender) Send() *Response {
	if p.err != nil {
		return &Response{err: p.err}
	}
	return send(p.info, p.req, p.cancel)
}

func (p *Sender) Abort() {
	p.cancel()
}

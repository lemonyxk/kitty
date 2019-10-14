package lemo

import (
	"github.com/Lemo-yxk/tire"
)

type HttpServer struct {
	IgnoreCase bool
	Router     *tire.Tire
	OnError    ErrorFunction

	group *httpServerGroup
	route *httpServerRoute
}

func (h *HttpServer) Ready() {

}

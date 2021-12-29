/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-13 19:33
**/

package http

type Protobuf struct {
	bts []byte
}

func (p *Protobuf) Bytes() []byte {
	return p.bts
}

/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:33
**/

package http

import (
	"encoding/xml"
)

type Xml struct {
	bts []byte
}

func (j *Xml) Reset(data any) error {
	var bts, err = xml.Marshal(data)
	if err != nil {
		return err
	}
	j.bts = bts
	return err
}

func (j *Xml) Bytes() []byte {
	return j.bts
}

func (j *Xml) String() string {
	return string(j.bts)
}

func (j *Xml) Decode(v any) error {
	return xml.Unmarshal(j.bts, v)
}

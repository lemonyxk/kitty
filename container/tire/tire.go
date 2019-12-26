/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-12-26 19:12
**/

package tire

import (
	"fmt"
	"strings"
)

const FH uint8 = 58
const XG uint8 = 47
const SC uint8 = 255
const WH uint8 = 63

type Tire struct {
	children      *[SC]*Tire
	char          byte
	childrenCount uint8
	Keys          []string
	Path          []byte
	Data          interface{}
}

func (t *Tire) ParseParams(pathBytes []byte) []string {

	if len(t.Keys) == 0 {
		return nil
	}

	var pathArray = strings.Split(string(pathBytes), "/")

	var res []string

	var i = 1

	var bLen = len(t.Path) - 1

	for index, value := range t.Path {

		if value == XG && index != bLen && t.Path[index+1] == FH {
			res = append(res, pathArray[i])
			i++
			continue
		}

		if value == XG {
			i++
		}
	}

	return res
}

func (t *Tire) Insert(path string, data interface{}) {

	if path == "" {
		return
	}

	var pathBytes = []byte(path)

	if pathBytes[0] != XG {
		return
	}

	// include ?
	if strings.Index(path, string(WH)) != -1 {
		panic(fmt.Sprintf("%s is include [%s]", path, string(WH)))
	}

	// repeat router
	if h := getFormatValue(t, formatPath(pathBytes)); h != nil {
		panic(fmt.Sprintf("%s is conflict with %s", path, h.Path))
	}

	var t1 = t

	var k []byte

	var ka []string

	var s = true

	for index := range pathBytes {

		var c = pathBytes[index]

		// if c > SC {
		//	panic(fmt.Sprintf("%s include special characters %s", path, string(c)))
		// }

		if c == FH && (index != 0 && pathBytes[index-1] == XG) {
			s = false
			if index == len(pathBytes)-1 || (index != 0 && pathBytes[index+1] == XG) {
				panic(fmt.Sprintf("%s is invalid, after [:] do not have any var", path))
			}
			continue
		}

		if s == false {
			k = append(k, c)
			if index == len(pathBytes)-1 || (index != 0 && pathBytes[index+1] == XG) {
				c = FH
			} else {
				continue
			}
		}

		if s == false {
			c = FH
		}

		if k != nil {
			ka = append(ka, string(k))
		}

		var p *Tire

		if t.children == nil {
			t.children = &[SC]*Tire{}
		}

		if t.children[c] != nil {
			p = t.children[c]
		} else {
			p = new(Tire)
			p.children = &[SC]*Tire{}
			p.char = c
		}

		if index == len(pathBytes)-1 {
			p.Keys = ka
			p.Path = pathBytes
			p.Data = data
		}

		t.children[c] = p
		t.childrenCount++

		t = p
		k = nil
		s = true

	}

	t = t1

}

func formatPath(pathBytes []byte) []byte {

	if pathBytes == nil || len(pathBytes) == 0 {
		return nil
	}

	if pathBytes[0] != XG {
		return nil
	}

	if len(pathBytes) == 1 {
		return []byte{XG}
	}

	var res []byte

	var s = true

	for index := range pathBytes {
		var c = pathBytes[index]

		if c == FH && pathBytes[index-1] == XG {
			res = append(res, c)
			s = false
			continue
		}

		if c == XG {
			s = true
		}

		if s == true {
			res = append(res, pathBytes[index])
		}

	}

	return res
}

func getFormatValue(t *Tire, pathBytes []byte) *Tire {

	var n = t.children

	if t.childrenCount == 0 {
		return nil
	}

	for index := range pathBytes {

		var c = pathBytes[index]

		if n[c] == nil {
			return nil
		}

		if n[c].char != 0 {

			if index == len(pathBytes)-1 && n[c].Path != nil {
				return n[c]
			}

			n = n[c].children
		}

	}

	return nil

}

func (t *Tire) GetValue(pathBytes []byte) *Tire {

	var n = t.children

	if t.childrenCount == 0 {
		return nil
	}

	var bLen = len(pathBytes) - 1

	var f = true

	for index := range pathBytes {

		// c == : ?

		var c = pathBytes[index]

		if c == FH {

			// is the latest char ?
			if index == bLen {

				if n[FH] != nil && n[FH].Path != nil {
					return n[FH]
				}

				return nil
			}

			continue
		}

		if n[c] == nil || f == false {

			f = false

			// is /
			if c == XG {

				// is the latest char ?
				if index == bLen {

					if n[FH] == nil {
						return nil
					}

					if n[FH].children[XG] == nil {
						return nil
					}

					if n[FH].children[XG].Path == nil {
						return nil
					}

					// remove / index
					return n[FH].children[XG]
				}

				// not the latest char
				// return nil
				if n[FH] == nil {
					return nil
				}

				// not children return nil
				if n[FH].children[XG] == nil {
					return nil
				}

				// reset n
				n = n[FH].children[XG].children

				f = true

				continue

			}

			// is the latest char ?
			if index == bLen {
				if n[FH] != nil && n[FH].Path != nil {
					return n[FH]
				}

				return nil
			}

			if n[FH] == nil {
				return nil
			}

			continue
		}

		if n[c].char != 0 {
			if index == bLen && n[c].Path != nil {
				return n[c]
			}

			n = n[c].children

			f = true
		}
	}

	return nil
}

func fn(node *Tire, res *[]*Tire) {
	if node == nil {
		return
	}
	for i := 0; i < len(node.children); i++ {
		if node.children[i] != nil {
			if node.children[i].Data != nil {
				*res = append(*res, node.children[i])
			}
			fn(node.children[i], res)
		}
	}
}

func (t *Tire) GetAllValue() []*Tire {
	var res []*Tire
	fn(t, &res)
	return res
}

package ws

import (
	"fmt"
)

const FH uint8 = 58
const XG uint8 = 47
const SC uint8 = 127

type tire struct {
	children      [SC]*tire
	char          byte
	childrenCount uint8
	keys          []string
	values        []string
	path          string
	data          interface{}
}

func (t *tire) Insert(path string, data interface{}) {

	if path == "" {
		return
	}

	var pathBytes = []byte(path)

	if pathBytes[0] != XG {
		return
	}

	// 是否重复
	if h := getFormatValue(t, formatPath(pathBytes)); h != nil {
		panic(fmt.Sprintf("%s is conflict with %s", path, h.path))
	}

	var t1 = t

	var k []byte

	var ka []string

	var s = true

	for index := range pathBytes {

		var c = pathBytes[index]

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

		var p *tire

		if t.children[c] != nil {
			p = t.children[c]
		} else {

			p = new(tire)
			p.children = [SC]*tire{}
			p.char = c
		}

		if index == len(pathBytes)-1 {
			p.keys = ka
			p.path = path
			p.data = data
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

	if pathBytes == nil {
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

func getFormatValue(t *tire, pathBytes []byte) *tire {

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

			if index == len(pathBytes)-1 && n[c].path != "" {
				return n[c]
			}

			n = n[c].children
		}

	}

	return nil

}

func (t *tire) GetValue(pathBytes []byte) *tire {

	var n = t.children

	if t.childrenCount == 0 {
		return nil
	}

	var va []string

	var s []byte

	var bLen = len(pathBytes) - 1

	for index := range pathBytes {

		var c = pathBytes[index]

		// c == : ?
		if c == FH {

			s = append(s, c)

			// is the latest char ?
			if index == bLen {
				if n[FH] != nil && n[FH].path != "" {

					if s != nil && len(s) > 0 {
						va = append(va, string(s))
					}

					n[FH].values = va
					return n[FH]
				}
				return nil
			}

			continue
		}

		if n[c] == nil {

			// is /
			if c == XG {

				if s != nil && len(s) > 0 {
					va = append(va, string(s))
				}

				// is the latest char ?
				if index == bLen {
					if n[FH] == nil {
						return nil
					}

					if n[FH].children[XG] == nil {
						return nil
					}

					n[FH].children[XG].values = va

					return n[FH].children[XG]
				}

				s = s[0:0]
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

				continue

			}

			s = append(s, c)

			// is the latest char ?
			if index == bLen {
				if n[FH] != nil && n[FH].path != "" {

					if s != nil && len(s) > 0 {
						va = append(va, string(s))
					}

					n[FH].values = va
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

			if index == bLen && n[c].path != "" {
				n[c].values = va
				return n[c]
			}

			n = n[c].children
		}

	}

	return nil

}

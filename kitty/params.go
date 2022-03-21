/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2020-07-19 18:09
**/

package kitty

type Params struct {
	Keys   []string
	Values []string
}

func (ps Params) ByName(name string) string {
	for i := 0; i < len(ps.Keys); i++ {
		if ps.Keys[i] == name {
			return ps.Values[i]
		}
	}
	return ""
}

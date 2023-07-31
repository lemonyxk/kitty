/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2020-07-19 18:09
**/

package socket

type Params map[string]string

func (ps Params) Get(name string) string {
	return ps[name]
}

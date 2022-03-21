/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:34
**/

package kitty

type Logger interface {
	Errorf(format string, args ...any)
	Warningf(format string, args ...any)
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
}

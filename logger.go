/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-13 19:34
**/

package kitty

type Logger interface {
	Errorf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

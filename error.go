/**
* @program: lemo
*
* @description:
*
* @author: Mr.Wang
*
* @create: 2019-09-25 20:37
**/

package lemo

func Error(err error) func() error {
	return func() error {
		return err
	}
}

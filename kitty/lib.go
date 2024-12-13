/**
* @program: kitty
*
* @create: 2024-12-14 00:51
**/

package kitty

import "unsafe"

type ef struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

func unpackEFace(obj interface{}) *ef {
	return (*ef)(unsafe.Pointer(&obj))
}

func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}
	return unpackEFace(obj).data == nil
}

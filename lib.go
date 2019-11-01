/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-10-05 14:18
**/

package lemo

func ParseMessage(bts []byte) ([]byte, []byte) {

	var s, e int

	var l = len(bts)

	if l < 9 {
		return nil, nil
	}

	// 正序
	if bts[8] == 58 {

		s = 8

		for i, b := range bts {
			if b == 44 {
				e = i
				break
			}
		}

		if e == 0 {
			return bts[s+2 : l-2], nil
		}

		return bts[s+2 : e-1], bts[e+9 : l-2]

	} else {

		for i := l - 1; i >= 0; i-- {

			if bts[i] == 58 {
				s = i
			}

			if bts[i] == 44 {
				e = i
				break
			}
		}

		if s == 0 {
			return nil, nil
		}

		if e == 0 {
			return bts[s+2 : l-2], nil
		}

		return bts[s+2 : l-2], bts[9 : e-1]
	}
}

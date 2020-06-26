/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2020-01-06 00:00
**/

package utils

type ne int

const Net ne = iota

//
// func (n ne) ConvertTCPListener(l net.Listener) *net.TCPListener {
// 	tl, ok := l.(*net.TCPListener)
// 	if !ok {
// 		return nil
// 	}
// 	return tl
// }
//
// func (n ne) SaveFD(l net.Listener, fileName string) error {
//
// 	var tl = n.ConvertTCPListener(l)
// 	if tl == nil {
// 		return errors.New("type error")
// 	}
//
// 	f, err := tl.File()
// 	if err != nil {
// 		return err
// 	}
//
// 	var fdPath = filepath.Join(os.TempDir(), fileName)
// 	ff, err := os.Create(fdPath)
// 	defer func() { _ = ff.Close() }()
//
// 	if err != nil {
// 		return err
// 	}
//
// 	var fd = fmt.Sprintf("%v", f.Fd())
// 	var name = f.Name()
//
// 	_, err = ff.Write([]byte(fmt.Sprintf("%s %s", fd, name)))
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (n ne) GetFD(fileName string) (*os.File, error) {
//
// 	var fdPath = filepath.Join(os.TempDir(), fileName)
//
// 	ff, err := os.Open(fdPath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	bts, err := ioutil.ReadAll(ff)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var arr = strings.Split(Conv.BytesToString(bts), " ")
// 	if len(arr) != 2 {
// 		return nil, errors.New("bad data")
// 	}
//
// 	fdStr, err := strconv.Atoi(arr[0])
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var fd, name = uintptr(fdStr), arr[1]
//
// 	var f = os.NewFile(fd, name)
//
// 	return f, nil
// }

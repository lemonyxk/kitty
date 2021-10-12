/**
* @program: kitty
*
* @description:
*
* @author: lemo
*
* @create: 2021-05-21 17:36
**/

package client

type writer struct {
	total      int64
	current    int64
	onProgress func(p []byte, current int64, total int64)
	last       int64
	rate       int64
}

func (w *writer) Write(p []byte) (int, error) {
	n := len(p)
	w.current += int64(n)

	if w.total == 0 {
		w.onProgress(p, w.current, -1)
	} else {
		w.last += int64(n) * w.rate

		if w.last >= w.total {
			w.onProgress(p, w.current, w.total)
			w.last = w.last - w.total
		}
	}

	return n, nil
}

type Progress struct {
	rate     int64
	progress func(p []byte, current int64, total int64)
}

// Rate 0.01 - 100
func (p *Progress) Rate(rate float64) *Progress {
	if rate < 0.01 || rate > 100 {
		rate = 1
	}
	p.rate = int64(100 / rate)
	return p
}

func (p *Progress) OnProgress(fn func(p []byte, current int64, total int64)) *Progress {
	p.progress = fn
	return p
}

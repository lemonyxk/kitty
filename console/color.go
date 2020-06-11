/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-27 20:16
**/

package console

import (
	"sync"

	color2 "github.com/fatih/color"
)

// Attribute defines a single SGR Code
type Color int

var cache map[Color]*color2.Color
var mux sync.RWMutex

// Base attributes
const (
	Reset Color = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

// Foreground text colors
const (
	FgBlack Color = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// Foreground Hi-Intensity text colors
const (
	FgHiBlack Color = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

// Background text colors
const (
	BgBlack Color = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Hi-Intensity text colors
const (
	BgHiBlack Color = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

func getCache(color Color) *color2.Color {
	mux.Lock()
	defer mux.Unlock()
	if c, ok := cache[color]; ok {
		return c
	}
	return color2.New(color2.Attribute(color))
}

func (c Color) Println(v ...interface{}) {
	_, _ = getCache(c).Println(v...)
}

func (c Color) Printf(format string, v ...interface{}) {
	_, _ = getCache(c).Printf(format, v...)
}

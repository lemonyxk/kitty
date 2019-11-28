/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-27 20:39
**/

package console

import tab "github.com/jedib0t/go-pretty/table"

const Asc = tab.Asc
const Dec = tab.Dsc

type table struct {
	write tab.Writer
}

func NewTable() *table {
	return &table{tab.NewWriter()}
}

func (t *table) Header(v ...interface{}) *table {
	t.write.AppendHeader(v)
	return t
}

func (t *table) Row(v ...interface{}) *table {
	t.write.AppendRow(v)
	return t
}

func (t *table) Footer(v ...interface{}) *table {
	t.write.AppendFooter(v)
	return t
}

func (t *table) SortByName(name string, mode tab.SortMode) *table {
	t.write.SortBy([]tab.SortBy{{Name: name, Mode: mode}})
	return t
}

func (t *table) SortByNumber(number int, mode tab.SortMode) *table {
	t.write.SortBy([]tab.SortBy{{Number: number, Mode: mode}})
	return t
}

func (t *table) Title(format string, v ...interface{}) *table {
	t.write.SetTitle(format, v...)
	return t
}

func (t *table) Render() string {
	return t.write.Render()
}

/**
* @program: kitty
*
* @description:
*
* @author: lemon
*
* @create: 2021-05-13 19:33
**/

package http

import (
	"github.com/bytedance/sonic/ast"
	"github.com/lemonyxk/kitty/json"
)

type Json struct {
	t    any
	bts  []byte
	root *Node
}

type Node struct {
	ast *ast.Node
}

func (n *Node) Get(path ...any) *Node {
	var node = n.ast.GetByPath(path...)
	return &Node{
		ast: node,
	}
}

func (n *Node) Int64() int64 {
	r, _ := n.ast.StrictInt64()
	return r
}

func (n *Node) String() string {
	r, _ := n.ast.StrictString()
	return r
}

func (n *Node) Float64() float64 {
	r, _ := n.ast.StrictFloat64()
	return r
}

func (n *Node) Bool() bool {
	r, _ := n.ast.Bool()
	return r
}

func (n *Node) Raw() string {
	r, _ := n.ast.Raw()
	return r
}

func (n *Node) Bytes() []byte {
	r, _ := n.ast.MarshalJSON()
	return r
}

func (n *Node) Array() []any {
	r, _ := n.ast.Array()
	return r
}

func (n *Node) Object() map[string]any {
	r, _ := n.ast.Map()
	return r
}

func (n *Node) Len() int {
	r, _ := n.ast.Len()
	return r
}

func (n *Node) Exists() bool {
	return n.ast.Exists()
}

func (j *Json) Get(path ...any) *Node {
	if j.root == nil {
		root, _ := json.Get(j.bts, path...)
		j.root = &Node{
			ast: &root,
		}
	}

	return j.root
}

func (j *Json) Reset(data any) error {
	var bts, err = json.Marshal(data)
	if err != nil {
		return err
	}
	j.bts = bts
	return err
}

func (j *Json) Bytes() []byte {
	return j.bts
}

func (j *Json) String() string {
	return string(j.bts)
}

func (j *Json) Decode(v any) error {
	j.t = v
	return json.Unmarshal(j.bts, v)
}

func (j *Json) Validate(v any) error {
	j.t = v
	return NewValidator[any]().From(j.bts).Bind(v)
}

package frame

import (
	"strings"
)

type tree struct {
	month string //方法
	root  *node  //根节点
}

type trees []tree

type nodeType uint8

const (
	static   nodeType = iota //默认的静态节点
	root                     //根节点
	param                    //带参数的节点
	catchAll                 //*
)

//节点
type node struct {
	//当前路径
	path string
	//指引 字节点的前面几个字母
	indices string
	//子节点
	children []*node
	//函数条
	hfuncs HandlerFuncs
	// 优先级 越大越nb
	priority uint32
	// 节点类型，包括static, root, param
	// static: 静态节点（默认）
	// root: 树的根节点
	// param: 参数节点
	nType nodeType
	// 节点是否是参数节点或通配
	wildChild bool
	// 完整路径
	fullPath string
}

//找到对应方法的根节点
func (t trees) getMonth(m string) *node {
	for _, v := range t {
		if v.month == m {
			return v.root
		}
	}
	return nil
}

//实现radix树
// /user/:name/...
func (n *node) addRoute(path string, handlers HandlerFuncs) {
	//通配符后不配拥有子节点
	if n.nType == catchAll {
		panic("this uri is catchAll")
	}

	//每进来一个  优先级就加一
	n.priority++
	l := len(n.path) //新来的路径长
	//啥也没有
	if len(n.children) == 0 && len(n.path) == 0 {
		n.path = path
		n.nType = getType(path)
		n.wildChild = isWildChild(path)
		n.hfuncs = handlers
		n.fullPath = path
		return
	}

	//匹配的数
	i := longestCommonPrefix(n.path, path)
	//通配节点

	if i < len(path) {
		//通配节点后不允许有其它节点
		if path[i] == '*' {
			panic("the\t" + path + "\tis catch all")
		}
	}
	//如果是参数节点 或者传进来的是个参数节点

	if n.nType == param || (i < len(path) && path[i] == ':') {
		n_section := strings.Split(n.path[i:], "/")
		p_section := strings.Split(path[i:], "/")
		//只有长度不同时，才判断
		if len(n_section) == len(p_section) {

			flag := false
			//遍历找不同
			for i := 0; i < min(len(n_section), len(p_section)); i++ {
				if p_section[i][0] == ':' || n_section[i][0] == ':' {
					continue
				}

				//如果有不同，就正常添加就好
				if p_section[i] != n_section[i] {
					flag = true
					break
				}
			}
			if !flag {

				panic("the\t" + path + "\tis exist")
			}
		}
	}

	//不是中间节点，完全匹配就是uri重复
	if i == len(path) && i == l {
		panic("the router exist\t " + path)
	}
	//if
	//存在交集
	if i < len(n.path) {
		//这个是自己变后
		copy_n := &node{
			path:      n.path[i:],
			indices:   n.indices,
			children:  n.children,
			hfuncs:    n.hfuncs,
			priority:  n.priority - 1,
			nType:     n.nType,
			wildChild: n.wildChild, //跟随父级
			fullPath:  n.fullPath,
		}

		//直接塞
		n.children = []*node{copy_n}
		n.path = n.path[:i]
		n.hfuncs = nil                                      //中间节点就不用方法了
		n.indices = string(copy_n.path[0])                  //现加它自己的
		n.fullPath = n.fullPath[:(len(n.fullPath) + i - l)] //到目前的路径
		n.nType = static
		n.wildChild = false

		//塞新来的这个
		//如果path对于n.path有不同的
		if i < len(path) {
			path = path[i:]
			//进行塞

			n.addChild(path, handlers)
			return

		} else {
			//n.path 包含了 path
			n.wildChild = copy_n.wildChild
			n.nType = copy_n.nType
			n.hfuncs = handlers
		}
		return
	} else { //path 包含n.path
		//如果没有子节点
		if len(n.indices) == 0 {

			n.addChild(path[i:], handlers)
			return
		}
		//遍历寻找

		for k := 0; k < len(n.indices); k++ {
			if n.indices[k] == path[i] {

				n.children[k].addRoute(path[i:], handlers)
				return
			}
		}
		//没有匹配的
		n.addChild(path[i:], handlers)
	}

}

//添加就只是添加就好
func (n *node) addChild(path string, handlers HandlerFuncs) {

	child := &node{
		path:      path,
		hfuncs:    handlers,
		priority:  n.priority - 1,
		nType:     getType(path),
		wildChild: isWildChild(path),
		fullPath:  n.fullPath + path,
	}
	n.children = append(n.children, child)
	n.indices = n.indices + string(path[0])

}

type Param struct {
	key   string
	value string
}

type Params []Param

//获取参数
func (p Params) Get(key string) (string, bool) {
	for _, v := range p {
		if v.key == key {
			return v.value, true
		}
	}
	return "", false
}

//获取参数byname
func (p Params) ByName(key string) string {
	v, _ := p.Get(key)
	return v
}

type nodeValue struct {
	handlers HandlerFuncs //函数
	fullPath string       //路径
	params   Params       //键值对
}

func (n *node) getValues(uri string) *nodeValue {

	switch n.nType {
	case static:
		index := 0
		for ; index < min(len(n.path), len(uri)); index++ {
			//如果有不同的 这个是静态路径
			if uri[index] != n.path[index] {
				//404
				return nil
			}
		}
		//如果遍历完了 并且要完全匹配
		if index == len(uri) && index == len(n.path) {
			return &nodeValue{
				handlers: n.hfuncs,
				fullPath: n.fullPath,
				params:   nil,
			}
		} else if index == len(uri) && index < len(n.path) {

			return nil
		}
		//如果没遍历完
		for i := 0; i < len(n.indices); i++ {
			if n.indices[i] == uri[index] {
				//如果找到了，，就返回
				if nv := n.children[i].getValues(uri[index:]); nv != nil {
					return nv
				}
				break
			}
		}
		//如果没找到匹配的静态
		//找动态
		for i := 0; i < len(n.indices); i++ {
			if n.indices[i] == '*' || n.indices[i] == ':' {
				return n.children[i].getValues(uri[index:])
			}
		}

		return nil
		//404
	case catchAll:
		//后面的全部匹配
		i := strings.Index(n.path, "*")
		//如果短了  404
		if i > len(uri) {
			return nil
		}
		return &nodeValue{
			handlers: n.hfuncs,
			fullPath: n.fullPath,
			params: []Param{
				{
					key:   n.path[i:],
					value: uri[i:],
				},
			},
		}
		//参数
	case param:
		var params []Param
		//找到各个参数的k v
		k_section := strings.Split(n.path, "/")
		v_section := strings.Split(uri, "/")
		//匹配 /user 和/user/ 这种
		if len(k_section) > len(v_section) {
			return nil
		}
		for k, v := range k_section {
			if i := strings.IndexRune(v, ':'); i != -1 {
				p := Param{
					key:   v[i+1:],
					value: v_section[k],
				}
				params = append(params, p)
				continue
			}
			if k_section[k] != v_section[k] {
				//404
				return nil
			}
		}
		//如果后面还有
		if len(v_section) > len(k_section) {
			path := ""
			for i := len(k_section); i < len(v_section); i++ {
				path = "/" + v_section[i]
			}
			if n.path[len(n.path)-1] == '/' {
				path = path[1:]
			}
			//遍历 寻找
			var nv *nodeValue
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == path[0] {
					nv = n.children[i].getValues(path)
				}

			}
			//未找到
			if nv == nil {
				return nil
			}
			//找到了
			nv.params = append(nv.params, params...)
			return nv
		}
		//后面没得
		return &nodeValue{
			handlers: n.hfuncs,
			fullPath: n.fullPath,
			params:   params,
		}
	default:
		//不可能是root
		return nil

	}

}

//判断是不是参数节点
func isWildChild(str string) bool {

	return strings.ContainsRune(str, ':'|'*')

}

func getType(str string) nodeType {
	if isWildChild(str) {
		return param
	}
	if strings.ContainsRune(str, '*') {
		return catchAll
	}
	return static
}

//最大匹配数
func longestCommonPrefix(pre string, path string) int {
	longest := 0
	for ; longest < min(len(pre), len(path)); longest++ {
		if pre[longest] != path[longest] {
			return longest
		}
	}
	return longest
}

//获取较小的一个
func min(a int, b int) int {
	if a >= b {
		return b
	}
	return a
}

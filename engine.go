package frame

import (
	"log"
	"net/http"
	"strings"
	"sync"
)

//万金油
type H map[string]interface{}

type HandlerFuncs []HandlerFunc

//trees -> tree ->node -> path
// radix树 提高查找的效率，是典型的以空间换取时间的做法
type Engine struct {
	group groupRouter //组

	middle []HandlerFunc //存放使用use方法时创建的函数

	others []*Engine //其它的

	pool sync.Pool //设置变量池，便于gc（垃圾处理）
	ts   trees     //前缀树
}

//这里的
type groupRouter struct {
	basePath string
}

//初始化
func Default() *Engine {
	return &Engine{}
}

func (e *Engine) POST(url string, h ...HandlerFunc) {
	e.handle(http.MethodPost, url, h)
}
func (e *Engine) GET(url string, h ...HandlerFunc) {
	e.handle(http.MethodGet, url, h)
}
func (e *Engine) PUT(url string, h ...HandlerFunc) {
	e.handle(http.MethodPut, url, h)
}
func (e *Engine) Delete(url string, h ...HandlerFunc) {
	e.handle(http.MethodDelete, url, h)
}
func (e *Engine) Any(url string, h ...HandlerFunc) {
	e.POST(url, h...)
	e.PUT(url, h...)
	e.Delete(url, h...)
	e.GET(url, h...)
}

//注册函数
func (e *Engine) handle(m string, url string, handlers []HandlerFunc) {
	if handlers == nil {
		log.Panic("the func can't be empty ")
		return
	}
	//处理uri
	url = initUri(e, url)

	//处理handlers

	if len(e.middle) != 0 {
		handlers = append(handlers, e.middle...)
	}

	t := e.ts.getMonth(m)
	if t == nil {
		//如果没有就添加一个
		e.ts = append(e.ts, tree{
			month: m,
			root:  &node{nType: root},
		})
		t = e.ts.getMonth(m)
	}
	t.addRoute(url, handlers)
	//没错就开始运行
	startFunc(m, url, len(handlers))

}
func (e *Engine) Use(f HandlerFunc) {
	e.middle = append(e.middle, f)
}

//对静态文件
func (e *Engine) StaticFile(url, path string) {
	//不能是动态路由
	if strings.Contains(path, ":") || strings.Contains(path, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	//创建函数，绑定文件
	handler := func(c *Context) {
		http.ServeFile(c.ResponseWriter, c.Request, path)
	}
	//注册路由
	e.GET(url, handler)

}

//对组进行注册
func (e *Engine) Group(url string) *Engine {
	//初始化处理uri
	url = initUri(e, url)

	group := &Engine{
		group: groupRouter{
			basePath: url,
		},
	}
	e.others = append(e.others, group)
	return group
}

func (e *Engine) Run(port string) {
	if port[0] != ':' {
		port = ":" + port

	}
	http.Handle("/", e)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Panic("listenandserver error: " + err.Error())
	}

}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	month := r.Method
	url := r.URL.Path
	c := NewContext(w, r)

	//放进变量池
	e.pool.Put(c)

	//浏览器会自动的请求favicon.ico  避免这种
	if r.URL.Path == "/favicon.ico" {
		return
	}
	//进行寻找
	flag := start(e /*engine*/, month /*month*/, url /*uri*/, c /*context*/)

	if len(e.others) != 0 && !flag {
		for _, v := range e.others {
			flag = start(v /*engine*/, month /*month*/, url /*uri*/, c /*context*/)
			//这里找到了相应的函数就退出，不浪费时间
			if flag {
				break
			}

		}
	}
	//打印
	if !flag {
		recordResponse(c.Request, http.StatusNotFound)
	} else {

		recordResponse(c.Request, http.StatusOK)
	}
}

//开始运行
func start(e *Engine, m string, url string, c *Context) bool {

	h := e.ts.getMonth(m)
	if h == nil {
		return false
	}

	nodeV := h.getValues(url)

	//可能时 /user 匹配 /user/ 这种
	if (nodeV == nil || nodeV.handlers == nil) && url[len(url)-1] != '/' {
		url := url + "/"
		nodeV = h.getValues(url)
	}
	//中间节点不配拥有函数
	if nodeV == nil || nodeV.handlers == nil {
		return false
	}
	//动态路由中的参数
	if len(nodeV.params) != 0 {
		c.Params = append(c.Params, nodeV.params...)
	}
	c.funcs = nodeV.handlers
	c.index = -1 //设置初始为-1，函数条才能正常跑
	c.Next()

	return true
}

//初始化一下
func initUri(e *Engine, url string) string {
	//uri前没得“/” 就加一个
	if url[0] != '/' {
		url = "/" + url
	}
	if len(e.group.basePath) != 0 {
		url = e.group.basePath + url
	}

	return valid(url)

}
func valid(url string) string {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' || url[i] == '*' {
			//后面必须有参数
			if i+1 == len(url) || url[i] == '/' {
				panic("there are no parameters after " + string(url[i]))
			}
			//'*'前一定要用‘/’
			if url[i] == '*' && url[i-1] != '/' {
				panic("no / before " + url)
			}
		}
		//去除多余的‘/’
		if url[i] == '/' && i != 0 && url[i-1] == '/' {
			url = url[:i] + url[i+1:]

		}

	}
	//返回小写
	return strings.ToLower(url)

}

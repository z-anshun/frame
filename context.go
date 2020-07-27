package frame

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter

	query url.Values  //路由传参
	form  *url.Values //表单

	Keys   map[string]interface{} //上下文中保存的变量
	Params Params                 //路由参数

	funcs HandlerFuncs //方法
	index int          //第一几个方法

	e *Engine //注入的引擎实例
}

type HandlerFunc func(*Context)

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:        r,
		ResponseWriter: w,
		query:          getQuery(r),
		form:           getForm(r),
	}

}

func (c *Context) WriteString(str string) {
	c.ResponseWriter.Write([]byte(str))
}

func (c *Context) Abort() {

	c.index = 255 //调整为很大的一个数
}

func (c *Context) Next() {
	c.index++
	for ; c.index < len(c.funcs); c.index++ {
		c.funcs[c.index](c)
	}
}

//获取表单
func getForm(r *http.Request) *url.Values {
	//设置最大读取的
	//直接
	//究极万金油
	//可能根本没有传表单
	r.ParseMultipartForm(128)
	//panic("Request data too large")

	//获取form
	form := r.PostForm

	return &form
}

//获取url中的参数
func getQuery(r *http.Request) url.Values {
	return r.URL.Query()
}

//Query			PostForm			获取key对应的值，不存在为空字符串
//GetQuery		GetPostForm			多返回一个key是否存在的结果
//QueryArray	PostFormArray		获取key对应的数组，不存在返回一个空数组
//GetQueryArray	GetPostFormArray	多返回一个key是否存在的结果
//QueryMap		PostFormMap			获取key对应的map，不存在返回空map
//GetQueryMap	GetPostFormMap		多返回一个key是否存在的结果
//DefaultQuery	DefaultPostForm		key不存在的话，可以指定返回的默认值

//头部参数
func (c *Context) Header(key string, value string) {
	if value == "" {
		c.ResponseWriter.Header().Del(key)
		return
	}

	c.Request.Header.Set(key, value)
}

//对于get请求的query
func (c *Context) Query(key string) string {
	v, _ := c.GetQuery(key)
	return v
}

//获取url中的value
func (c *Context) GetQuery(key string) (value string, exist bool) {
	value = c.query.Get(key)
	if len(value) == 0 {
		exist = false
		return
	}
	exist = true
	return
}

//获取数组
func (c *Context) QueryArray(key string) []string {
	v, _ := c.GetQueryArray(key)
	return v
}

//是否获取请求的数组
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	v := c.query[key]
	exist := false
	if len(v) != 0 {
		exist = true

	}
	return v, exist
}

//是否获取到请求的map
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	return get(c.query, key)
}

func (c *Context) QueryMap(key string) map[string]string {
	v, _ := get(c.query, key)
	return v
}

//default请求
func (c *Context) DefaultQuery(key string, value string) string {
	if v, ok := c.GetQuery(key); ok {
		return v
	}
	return value
}

//对于表单请求
func (c *Context) PostParam(key string) string {
	v, _ := c.GetPostForm(key)
	return v
}

//获取表单的value
func (c *Context) GetPostForm(key string) (string, bool) {
	v := c.form.Get(key)
	exits := false
	if len(v) > 0 {
		exits = true
		return v, exits
	}

	return v, exits
}

//获取表单数组
func (c Context) PostFormArray(key string) []string {
	v, _ := c.GetPostFormArray(key)
	return v
}

//是否获取到表单数组
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	v := (*c.form)[key]
	exist := false
	if len(v) > 0 {
		exist = true
	}
	return v, exist
}

//是否获取表单到map数据
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	return get(*c.form, key)
}

func (c *Context) PostFormMap(key string) map[string]string {
	v, _ := get(*c.form, key)
	return v
}

//default表单的请求
func (c *Context) DefaultPostForm(key string, value string) string {
	if v, ok := c.GetPostForm(key); ok {
		return v
	}
	return value
}

//设置k-value
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

//获取k-value
func (c *Context) Get(key string) interface{} {
	return c.Keys[key]
}

//获取param
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

//获取cookie
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	//将QueryEscape转码的字符串还原
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

//设置cookie
//k v 最大时间 路径 域   是否只能通过https来传递此条cookie 是否只存在请求头中
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.ResponseWriter, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value), //转码 可以安全的用在URL查询里
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

//渲染html 数据 文件
func (c *Context) TempleHTMLFiles(obj interface{}, files ...string) error {
	parse, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}
	if err := parse.Execute(c.ResponseWriter, obj); err != nil {
		//渲染失败
		return err
	}
	return nil
}

//指定生成的文件
func (c *Context) TempleDelimsHTMLFiles(obj interface{}, left string, right string, files ...string) error {
	// 解析指定文件生成模板对象
	t, err := template.New("").Delims(left, right).ParseFiles(files...)
	if err != nil {
		return err
	}
	//渲染输出
	if err := t.Execute(c.ResponseWriter, obj); err != nil {
		return err
	}
	return nil

}

//websocket升级
func (c *Context) UpdateWs() *websocket.Conn {
	if c.Request.Method != http.MethodGet {
		return nil
	}
	upgrade := websocket.Upgrader{
		//实现跨域，怎么都是true
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrade.Upgrade(c.ResponseWriter, c.Request, nil)
	if err != nil {
		log.Println("websocket connect is failed,err:", err)
		return nil
	}
	return conn
}

//对于返回的信息
func statusOk(status int) bool {
	switch {
	case status >= 100 && status <= 199: //服务器禁止向此类客户端发送 1xx 响应
		return false
	case status == http.StatusNoContent: //204 未含消息
		return false
	case status == http.StatusNotModified: //304响应禁止包含消息体
		return false
	}
	return true
}

//写入status
func (c *Context) Status(code int) {
	c.ResponseWriter.WriteHeader(code)
}

//设置Content-Type
func (c *Context) SetContentType(value string) {
	c.ResponseWriter.Header().Set("Content-Type", value)
}

//返回json化的数剧
func (c Context) Json(code int, obj interface{}) {
	if !statusOk(code) {
		return
	}
	c.SetContentType("application/json; charset=utf-8")
	jsonObj, err := json.Marshal(obj)
	if err != nil {
		return
	}
	c.ResponseWriter.Write(jsonObj)
}

//?name[name]=as 真有人这么传参？
func get(m map[string][]string, key string) (map[string]string, bool) {
	value := make(map[string]string)
	exist := false

	for k, v := range m {
		if i := strings.IndexByte(k, '['); i >= 1 {
			if j := strings.IndexByte(k[i:], ']'); j > i {
				value[k[i:j]] = v[0]
				exist = true
			}
		}
	}
	return value, exist

}

package frame

import (
	"errors"
	"github.com/z-anshun/frame/binding"
	"strings"
)

//bind json类型的数据
func (c *Context) BindJson(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

//yaml
func (c *Context) BindYaml(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.YAML)
}

//xml
func (c *Context) BindXml(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.XML)
}

//msgpack
func (c *Context) BindMsgPack(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.MsgPack)
}

//获取元素
func (c *Context) Bind(obj interface{}) error {
	t := c.Request.Header.Get("Content-Type")
	if strings.Contains(t, "multipart/form-data") {
		t = "multipart/form-data"
	}
	b := binding.Default(c.Request.Method, t)
	return c.ShouldBindWith(obj, b)
}

//通过接口找到对应结构体的方法
func (c *Context) ShouldBindWith(obj interface{}, b binding.Bind) error {
	if err := validInterface(obj); err != nil {
		return err
	}

	return b.Bind(c.Request, obj)
}

//ValidateStruct可以接收任何类型的数据，它永远不会死机，即使配置不正确。
//如果接收到的类型不是结构，则应跳过任何验证，并且必须返回零。
//如果接收的类型是结构或指向结构的指针，则应执行验证。
//如果结构无效或验证本身失败，则应返回描述性错误。
//否则必须返回零。

func validInterface(i interface{}) error {
	if i == nil {
		return errors.New("this pointer cannot be null")
	}
	return nil
}

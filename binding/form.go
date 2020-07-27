package binding

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type formBinding struct{}
type formPostBinding struct{}
type formMultipartBinding struct{}

func (formBinding) Name() string {
	return "form"
}

//get中的表单
func (formBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	//最大读取32mb
	if err := req.ParseMultipartForm(32 * 1024 * 1024); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}
	b := req.Form

	return decodeFrom(b, obj)
}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

//原生form
func (formPostBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	b := req.Form

	return decodeFrom(b, obj)
}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

//使用表单上传文件
func (formMultipartBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseMultipartForm(128); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}

	b := req.PostForm

	return decodeFrom(b, obj)
}

func decodeFrom(b url.Values, obj interface{}) error {

	form := make(map[string]interface{})

	str := ""
	//获取第一个
	for k, v := range b {

		//正常的map
		if len(v) != 0 {
			form[k] = v[0]
			continue
		}
		//如果是那种都塞在头那种
		str = str + k
		if len(str) == 0 {
			break
		}
	}
	//获取到了正常的map
	if len(form) != 0 {
		m, err := json.Marshal(form)
		if err != nil {
			return err
		}
		//只能是字符串
		return decodeJSON(bytes.NewReader(m), obj)
	}

	v := strings.NewReader(str)

	return decodeJSON(v, obj)
}

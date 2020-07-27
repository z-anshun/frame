package binding

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

//
func (xmlBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return decodeXML(req.Body, obj)
}

func (xmlBinding) BindBody(body []byte, obj interface{}) error {
	return decodeXML(bytes.NewReader(body), obj)
}

func decodeXML(b io.Reader, obj interface{}) error {
	return xml.NewDecoder(b).Decode(obj) //使用decoder数据流中读取，不用缓存
}

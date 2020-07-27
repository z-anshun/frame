package binding

import (
	"bytes"
	"github.com/ugorji/go/codec"
	"io"
	"net/http"
)

type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj interface{}) error {
	return decodeMsgPack(req.Body, obj)
}

func (msgpackBinding) BindBody(body []byte, obj interface{}) error {
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj interface{}) error {
	//这里要引用一个外包来序列化msgpack文件
	cdc := new(codec.MsgpackHandle)
	return codec.NewDecoder(r, cdc).Decode(&obj)
}

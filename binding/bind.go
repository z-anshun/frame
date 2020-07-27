package binding

import "net/http"

const (
	MIMEJSON              = "application/json"                  //序列化后的 JSON 字符串 json数据格式
	MIMEHTML              = "text/html"                         //HTML格式
	MIMEXML               = "application/xml"                   //XML数据格式
	MIMEXML2              = "text/xml"                          //XML格式
	MIMEPlain             = "text/plain"                        //纯文本模式
	MIMEPOSTForm          = "application/x-www-form-urlencoded" //浏览器的原生form表单
	MIMEMultipartPOSTForm = "multipart/form-data"               //使用表单上传文件时
	MIMEPROTOBUF          = "application/x-protobuf"            //protobuf数据格式
	MIMEMSGPACK           = "application/x-msgpack"             //x-开头的方法标识这个类别还没有成为标准
	MIMEMSGPACK2          = "application/msgpack"               //二进制序列化数据格式
	MIMEYAML              = "application/x-yaml"                //yaml数据格式
)

//耿直bind
//创建三个接口，便于后面更改与处理
type Bind interface {
	Name() string
	Bind(r *http.Request, obj interface{}) error
}

type BindBody interface {
	Name() string
	BindBody([]byte, interface{}) error
}
type BindUri interface {
	Name() string
	BindUri(map[string][]string, interface{}) error
}

var (
	JSON = jsonBinding{}
	XML  = xmlBinding{}
	Form = formBinding{}
	//Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	//ProtoBuf      = protobufBinding{}
	MsgPack       = msgpackBinding{}
	YAML          = yamlBinding{}
	//Uri           = uriBinding{}
	//Header        = headerBinding{}
)

//返回对应类型的接口 给bind用的
func Default(method, contentType string) Bind {
	if method == "GET" {
		return Form
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	//case MIMEPROTOBUF:
	//	return ProtoBuf
	case MIMEMSGPACK, MIMEMSGPACK2:
		return MsgPack
	case MIMEYAML:
		return YAML
	case MIMEMultipartPOSTForm:
		return FormMultipart
	case MIMEPOSTForm:
		return FormPost
	default:
		return Form
	}
}

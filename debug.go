package frame

import (
	"fmt"
	"net/http"
	"time"
)

func recordResponse(r *http.Request, status uint) {
	//打印出来，，哪个访问
	fmt.Println("Request:\t", r.Method, "\t", time.Now().Format("2006-01-02 15:04:05"), "\tstatus:", status, "\t url:", r.URL.Path)
}

func startFunc(m string, url string, handlers int) {
	fmt.Println("router:\t", m, "\t", url, "\t(", handlers, "handlers)")
}

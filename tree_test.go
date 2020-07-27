package frame

import (
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {
	n := node{}
	n.addRoute("/ad/as", HandlerFuncs{})
	n.addRoute("/ad:name", HandlerFuncs{})
	n.addRoute("/as/123", nil)

	for _, v := range n.children {
		fmt.Println("path:", v.path, "\t", v.fullPath, "\t", v.priority)
		if len(v.children) != 0 {
			for _, v2 := range v.children {
				fmt.Println("path:", v2.path, "\t", v2.fullPath, "\t", v2.priority)
			}
		}
	}
	fmt.Println(n.getValues("/ad123"))
}



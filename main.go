package main

import(
	"github.com/mohamed-rafraf/k8s-auth-server/core"
)

func main() {
	err:=core.StartServer()
	if err!=nil {
		return
	}
}
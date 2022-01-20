package main

import (
	"github.com/snowyyj001/loumiao/gate"
)

func Hotfix_update_func() {
	gate.Fix_bug_1()
}

/*/
go build --buildmode=plugin -o hotfix.1.so hotfix.go
*/

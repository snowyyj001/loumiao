package lbase

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"strings"
)

// 获取常量对应的注释
func Ast(fileName string) map[int]string {
	names := map[int]string{}
	fset := token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	path, _ := filepath.Abs(fileName)
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err.Error())
		return names
	}
	nLen := len(f.Decls)
	for i := 0; i < nLen; i++ {
		constVal := 0
		decl, ok := f.Decls[i].(*ast.GenDecl)
		if ok && decl.Tok == token.CONST {
			for _, v1 := range decl.Specs {
				val, ok := v1.(*ast.ValueSpec)
				if ok {
					constVal++
					if len(val.Values) >= 1 {
						val, ok := val.Values[0].(*ast.BasicLit)
						if ok {
							constVal, _ = Int(val.Value)
						}
					}
					names[constVal] = strings.Trim(val.Comment.Text(), "\n")
				}
			}
		}
	}

	return names
}

func dump(typeof reflect.Type, valueof reflect.Value) {
	if typeof.Kind() == reflect.Struct {
		fmt.Print(fmt.Sprintf("struct: %s---> ", typeof.Name()))
		for i := 0; i < typeof.NumField(); i++ {
			value := valueof.Field(i).Interface()
			dump(reflect.TypeOf(value), reflect.ValueOf(value))
		}
		fmt.Print("\n")
	} else if typeof.Kind() == reflect.Slice {
		fmt.Print("slice--->")
		for i := 0; i < valueof.Len(); i++ {
			value := valueof.Index(i).Interface()
			dump(reflect.TypeOf(value), reflect.ValueOf(value))
		}
		fmt.Print("\n")
	} else if typeof.Kind() == reflect.Array {
		fmt.Print("array--->")
		for i := 0; i < valueof.Len(); i++ {
			value := valueof.Index(i).Interface()
			dump(reflect.TypeOf(value), reflect.ValueOf(value))
		}
		fmt.Print("\n")
	} else if typeof.Kind() == reflect.Map {
		fmt.Print("map--->")
		for i := 0; i < valueof.Len(); i++ {
			value := valueof.Index(i).Interface()
			dump(reflect.TypeOf(value), reflect.ValueOf(value))
		}
		fmt.Print("\n")
	} else if typeof.Kind() == reflect.Ptr {
		typeodval := typeof.Elem()
		if typeodval.Kind() == reflect.Struct {
			dump(typeodval, valueof.Elem())
		} else {
			fmt.Print(fmt.Sprintf("%v ", valueof.Elem().Interface()))
		}
	} else {
		fmt.Print(fmt.Sprintf("%v ", valueof.Interface()))
	}
}

// dump信息,调试用，发布环境不要用
func Dump(name string, value interface{}) {
	fmt.Println("Dump=============>", name)
	typeof := reflect.TypeOf(value)
	valueof := reflect.ValueOf(value)
	dump(typeof, valueof)
}

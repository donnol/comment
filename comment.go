package comment

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
)

// Field 字段
type Field struct {
	Name    string // 名字
	Kind    string // 类型 TODO
	Comment string // 注释
}

// Struct 结构体
type Struct struct {
	Name    string  // 名字
	Comment string  // 注释
	Fields  []Field // 结构体字段
}

// MakeStruct 新建结构体
func MakeStruct() Struct {
	return Struct{
		Fields: make([]Field, 0),
	}
}

// ResolveStruct 解析结构体
func ResolveStruct(value interface{}) (Struct, error) {
	s := MakeStruct()

	var r = make(map[string]string)
	var f = make(map[string]string)

	refType := reflect.TypeOf(value)
	if refType.Kind() == reflect.Ptr { // 指针
		refType = refType.Elem()
	}
	if refType.Kind() != reflect.Struct {
		return s, fmt.Errorf("bad value type , type is %v", refType.Kind())
	}
	structName := refType.PkgPath() + "." + refType.Name()
	s.Name = structName

	if err := collectStructComment(refType, r, f); err != nil {
		return s, err
	}
	s.Comment = r[structName]
	for k, v := range f {
		s.Fields = append(s.Fields, Field{
			Name:    k,
			Comment: v,
		})
	}

	return s, nil
}

// collectStructComment 收集结构体的注释
func collectStructComment(refType reflect.Type, structCommentMap, fieldCommentMap map[string]string) error {
	// 获取结构体路径和名称
	var structName string
	structName = refType.PkgPath() + "." + refType.Name()

	// 解析
	if err := resolve(structName, structCommentMap, fieldCommentMap); err != nil {
		return fmt.Errorf("resolve output failed, error is %v", err)
	}

	// 内嵌结构体
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		if field.Anonymous { // 匿名
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr { // 指针
				fieldType = fieldType.Elem()
			}
			collectStructComment(fieldType, structCommentMap, fieldCommentMap)
		}
	}

	return nil
}

const (
	structStart = "type"
	structEnd   = "}"
	fieldSep    = " "
	commentSep  = "//"
)

// 返回结构体注释，字段名注释映射和错误
func resolve(structName string, structCommentMap, fieldCommentMap map[string]string) error {
	// 运行go doc命令
	cmd := exec.Command("go", "doc", structName)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var isEnd bool
	buf := bytes.NewBuffer(output)
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if strings.TrimSpace(line) == structEnd {
			isEnd = true
			continue
		}

		var comment string
		pieceList := strings.Split(line, commentSep)
		if !isEnd {
			keyList := strings.Split(strings.TrimSpace(pieceList[0]), fieldSep)
			if len(keyList) == 1 { // 匿名结构体
				continue
			}
			key := keyList[0]
			if key == structStart {
				continue
			}
			if len(pieceList) == 2 {
				comment = strings.TrimSpace(pieceList[1])
			}

			fieldCommentMap[key] = comment
		} else {
			pieceList = strings.Split(strings.TrimSpace(pieceList[0]), fieldSep)
			comment = strings.TrimSpace(pieceList[len(pieceList)-1])
			if comment != "" {
				if _, ok := structCommentMap[structName]; !ok {
					structCommentMap[structName] = comment
				}
			}
		}
	}

	return nil
}

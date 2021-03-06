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
	reflect.StructField        // 内嵌反射结构体字段类型
	Comment             string // 注释
	Struct              Struct // 字段的类型是其它结构体
}

// Struct 结构体
type Struct struct {
	Name        string       // 名字
	Comment     string       // 注释
	Description string       // 描述
	Type        reflect.Type // 反射类型
	Fields      []Field      // 结构体字段
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

	var refType reflect.Type
	if v, ok := value.(reflect.Type); ok {
		refType = v
	} else {
		refType = reflect.TypeOf(value)
	}
	s.Type = refType

	if refType.Kind() == reflect.Ptr { // 指针
		refType = refType.Elem()
	}
	if refType.Kind() != reflect.Struct {
		return s, fmt.Errorf("bad value type , type is %v", refType.Kind())
	}
	structName := refType.PkgPath() + "." + refType.Name()
	s.Name = structName

	if err := collectStructComment(refType, &s); err != nil {
		return s, err
	}

	return s, nil
}

// collectStructComment 收集结构体的注释
func collectStructComment(refType reflect.Type, s *Struct) error {
	// 解析-获取结构体注释
	var r = make(map[string]string)
	var f = make(map[string]string)
	var err error
	if r, f, err = resolve(s.Name); err != nil {
		return fmt.Errorf("resolve output failed, error is %v", err)
	}
	s.Comment = r[commentKey]
	s.Description = r[descriptionKey]

	// 内嵌结构体
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)

		sf := Field{
			StructField: field,
			Comment:     f[field.Name],
		}

		fieldType := field.Type
		if field.Anonymous { // 匿名
			sf.Struct, err = ResolveStruct(fieldType)
			if err != nil {
				return err
			}
		}
		// 非匿名结构体类型
		if fieldType.Kind() == reflect.Ptr ||
			fieldType.Kind() == reflect.Slice ||
			fieldType.Kind() == reflect.Map ||
			fieldType.Kind() == reflect.Chan ||
			fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			sf.Struct, err = ResolveStruct(fieldType)
			if err != nil {
				return err
			}
		}

		s.Fields = append(s.Fields, sf)
	}

	return nil
}

const (
	structStart    = "type"
	structEnd      = "}"
	fieldSep       = " "
	commentSep     = "//"
	commentKey     = "comment"
	descriptionKey = "description"
)

// 返回结构体注释，字段名注释映射和错误
func resolve(structName string) (map[string]string, map[string]string, error) {
	var structCommentMap = make(map[string]string)
	var fieldCommentMap = make(map[string]string)

	// 运行go doc命令
	cmd := exec.Command("go", "doc", structName)
	output, err := cmd.Output()
	if err != nil {
		return structCommentMap, fieldCommentMap, fmt.Errorf("go doc failed, struct name is %s, error is %v", structName, err)
	}

	var isEnd bool
	buf := bytes.NewBuffer(output)
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return structCommentMap, fieldCommentMap, fmt.Errorf("go doc failed, struct name is %s, error is %v", structName, err)
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
			if len(pieceList) > 1 {
				comment = strings.TrimSpace(pieceList[1])
				if _, ok := structCommentMap[commentKey]; !ok {
					structCommentMap[commentKey] = comment
				}
				if len(pieceList) > 2 {
					desc := strings.TrimSpace(pieceList[2])
					if _, ok := structCommentMap[descriptionKey]; !ok {
						structCommentMap[descriptionKey] = desc
					}
				}
			}
		}
	}

	return structCommentMap, fieldCommentMap, nil
}

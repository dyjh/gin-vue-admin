package model

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const maxOrderFoodTableNameLength = 30

// TestPersistenceModelConventions 防止业务模型重新引入长表名、模糊字段或缺失注释。
func TestPersistenceModelConventions(t *testing.T) {
	t.Helper()

	files, err := parser.ParseDir(token.NewFileSet(), ".", func(info fs.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse models: %v", err)
	}

	packageFiles := files["model"].Files
	tableNames := make(map[string]string)
	structs := make(map[string]struct{})

	for _, file := range packageFiles {
		inspectPersistenceStructs(t, file, structs)
		inspectTableNameMethods(t, file, tableNames)
	}

	for structName := range structs {
		if _, exists := tableNames[structName]; !exists {
			t.Errorf("%s must declare a documented TableName method", structName)
		}
	}
}

// inspectPersistenceStructs 校验模型、字段注释以及持久化字段的完整 GORM 标签。
func inspectPersistenceStructs(t *testing.T, file *ast.File, structs map[string]struct{}) {
	t.Helper()

	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structName := typeSpec.Name.Name
			structs[structName] = struct{}{}
			if general.Doc == nil && typeSpec.Doc == nil {
				t.Errorf("%s must have a type comment", structName)
			}

			for _, field := range structType.Fields.List {
				inspectPersistenceField(t, structName, field)
			}
		}
	}
}

// inspectPersistenceField 校验单个字段的同行说明和数据库列元数据。
func inspectPersistenceField(t *testing.T, structName string, field *ast.Field) {
	t.Helper()

	if field.Comment == nil {
		t.Errorf("%s field must have an inline comment", structName)
	}
	if len(field.Names) == 0 {
		return
	}
	fieldName := field.Names[0].Name
	if field.Tag == nil {
		t.Errorf("%s.%s must declare json and gorm tags", structName, fieldName)
		return
	}

	rawTag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		t.Errorf("%s.%s has invalid struct tags: %v", structName, fieldName, err)
		return
	}
	tag := reflect.StructTag(rawTag)
	if _, exists := tag.Lookup("json"); !exists {
		t.Errorf("%s.%s must declare a json tag", structName, fieldName)
	}
	gormTag, exists := tag.Lookup("gorm")
	if !exists {
		t.Errorf("%s.%s must declare a gorm tag", structName, fieldName)
		return
	}

	// 关联字段和显式忽略字段不对应独立数据库列，只要求关系定义与同行说明。
	if gormTag == "-" || strings.Contains(gormTag, "foreignKey:") || strings.Contains(gormTag, "many2many:") {
		return
	}
	for _, required := range []string{"column:", "type:", "comment:"} {
		if !strings.Contains(gormTag, required) {
			t.Errorf("%s.%s gorm tag must include %s", structName, fieldName, required)
		}
	}
	if !strings.Contains(gormTag, "not null") && !strings.Contains(gormTag, "default:null") {
		t.Errorf("%s.%s gorm tag must declare nullability", structName, fieldName)
	}
	if strings.Contains(gormTag, "order_food") {
		t.Errorf("%s.%s still uses a legacy long index or table name", structName, fieldName)
	}
}

// inspectTableNameMethods 校验表名注释、长度、前缀和唯一性。
func inspectTableNameMethods(t *testing.T, file *ast.File, tableNames map[string]string) {
	t.Helper()

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "TableName" || function.Recv == nil {
			continue
		}
		if function.Doc == nil {
			t.Errorf("TableName method must have a method comment")
		}

		receiver, ok := function.Recv.List[0].Type.(*ast.Ident)
		if !ok {
			t.Errorf("TableName receiver must be a concrete model")
			continue
		}
		if len(function.Body.List) != 1 {
			t.Errorf("%s.TableName must directly return a table name", receiver.Name)
			continue
		}
		returnStatement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(returnStatement.Results) != 1 {
			t.Errorf("%s.TableName must directly return a table name", receiver.Name)
			continue
		}
		literal, ok := returnStatement.Results[0].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Errorf("%s.TableName must return a string literal", receiver.Name)
			continue
		}
		tableName, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Errorf("%s.TableName contains an invalid string literal", receiver.Name)
			continue
		}
		if !strings.HasPrefix(tableName, "of_") {
			t.Errorf("%s table %q must use the of_ prefix", receiver.Name, tableName)
		}
		if len(tableName) > maxOrderFoodTableNameLength {
			t.Errorf("%s table %q exceeds %d characters", receiver.Name, tableName, maxOrderFoodTableNameLength)
		}
		for previous, existingTableName := range tableNames {
			if existingTableName == tableName {
				t.Errorf("%s and %s use the same table %q", previous, receiver.Name, tableName)
			}
		}
		tableNames[receiver.Name] = tableName
	}
}

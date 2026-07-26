package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var orderFoodCommentRoots = []string{
	"api/v1/orderfood",
	"front/api",
	"service/orderfood",
	"front/service",
	"model/orderfood",
	"front/request",
	"front/response",
}

// TestOrderFoodCommentConventions 防止订单餐模块的代码注释和 Swagger 契约回退。
func TestOrderFoodCommentConventions(t *testing.T) {
	t.Helper()

	fileSet := token.NewFileSet()
	for _, root := range orderFoodCommentRoots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, "_test.go") {
				return nil
			}

			file, parseErr := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if parseErr != nil {
				return parseErr
			}
			checkOrderFoodFileComments(t, fileSet, path, file)
			return nil
		})
		if err != nil {
			t.Fatalf("检查 %s 注释失败: %v", root, err)
		}
	}
}

func checkOrderFoodFileComments(
	t *testing.T,
	fileSet *token.FileSet,
	path string,
	file *ast.File,
) {
	t.Helper()

	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if typed.Name.IsExported() {
				requireNamedGoDoc(t, fileSet, path, typed.Name.Name, typed.Pos(), typed.Doc)
			}
			if typed.Name.IsExported() && strings.Contains(path, "/api/") &&
				isOrderFoodGinHandler(typed) {
				checkOrderFoodSwagger(t, fileSet, path, typed)
			}
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			for _, rawSpec := range typed.Specs {
				typeSpec := rawSpec.(*ast.TypeSpec)
				if !typeSpec.Name.IsExported() {
					continue
				}
				doc := typeSpec.Doc
				if doc == nil {
					doc = typed.Doc
				}
				requireNamedGoDoc(t, fileSet, path, typeSpec.Name.Name, typeSpec.Pos(), doc)
				checkOrderFoodStructFields(t, fileSet, path, typeSpec)
			}
		}
	}
}

func requireNamedGoDoc(
	t *testing.T,
	fileSet *token.FileSet,
	path string,
	name string,
	position token.Pos,
	doc *ast.CommentGroup,
) {
	t.Helper()

	if doc != nil && strings.HasPrefix(strings.TrimSpace(doc.Text()), name+" ") {
		return
	}
	t.Errorf(
		"%s: %s 缺少以名称开头的用途注释",
		sourcePosition(fileSet, path, position),
		name,
	)
}

func checkOrderFoodStructFields(
	t *testing.T,
	fileSet *token.FileSet,
	path string,
	typeSpec *ast.TypeSpec,
) {
	t.Helper()

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			if field.Comment == nil {
				t.Errorf(
					"%s: %s 的嵌入字段缺少同行注释",
					sourcePosition(fileSet, path, field.Pos()),
					typeSpec.Name.Name,
				)
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() && field.Comment == nil {
				t.Errorf(
					"%s: %s.%s 缺少字段同行注释",
					sourcePosition(fileSet, path, field.Pos()),
					typeSpec.Name.Name,
					name.Name,
				)
			}
		}
	}
}

func checkOrderFoodSwagger(
	t *testing.T,
	fileSet *token.FileSet,
	path string,
	function *ast.FuncDecl,
) {
	t.Helper()

	doc := ""
	if function.Doc != nil {
		doc = function.Doc.Text()
	}
	for _, annotation := range []string{
		"@Tags",
		"@Summary",
		"@Security",
		"@accept",
		"@Produce",
		"@Success",
		"@Router",
	} {
		if !strings.Contains(doc, annotation) {
			t.Errorf(
				"%s: %s 缺少 Swagger 注解 %s",
				sourcePosition(fileSet, path, function.Pos()),
				function.Name.Name,
				annotation,
			)
		}
	}
	if strings.Contains(doc, "data=object") ||
		strings.Contains(doc, "@Success 200 {object} response.Response\n") {
		t.Errorf(
			"%s: %s 的成功响应必须声明具体数据模型",
			sourcePosition(fileSet, path, function.Pos()),
			function.Name.Name,
		)
	}
}

func isOrderFoodGinHandler(function *ast.FuncDecl) bool {
	if function.Recv == nil || function.Type.Params == nil ||
		len(function.Type.Params.List) != 1 {
		return false
	}
	star, ok := function.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Context"
}

func sourcePosition(
	fileSet *token.FileSet,
	path string,
	position token.Pos,
) string {
	location := fileSet.Position(position)
	return fmt.Sprintf("%s:%d", path, location.Line)
}

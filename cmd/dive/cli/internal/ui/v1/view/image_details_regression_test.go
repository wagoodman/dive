package view

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func imageDetailsSourcePath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	return filepath.Join(filepath.Dir(thisFile), "image_details.go")
}

func imageDetailsAST(t *testing.T) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, imageDetailsSourcePath(t), nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("unable to parse image_details.go: %+v", err)
	}
	return file
}

func methodDeclByName(file *ast.File, recvType, methodName string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name == nil || fn.Name.Name != methodName || len(fn.Recv.List) != 1 {
			continue
		}
		star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		ident, ok := star.X.(*ast.Ident)
		if !ok {
			continue
		}
		if ident.Name == recvType {
			return fn
		}
	}
	return nil
}

func TestImageDetailsRenderPotentialWastedSpaceUsesRemovableBytes(t *testing.T) {
	file := imageDetailsAST(t)
	renderMethod := methodDeclByName(file, "ImageDetails", "Render")
	if renderMethod == nil {
		t.Fatal("missing ImageDetails.Render declaration")
	}

	found := false
	ast.Inspect(renderMethod.Body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || assign.Tok != token.ADD_ASSIGN || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}

		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != "wastedSpace" {
			return true
		}

		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil || sel.Sel.Name != "WastedSize" {
			return true
		}

		found = true
		return false
	})

	if !found {
		t.Fatal("ImageDetails.Render should sum removable bytes via data.WastedSize()")
	}
}

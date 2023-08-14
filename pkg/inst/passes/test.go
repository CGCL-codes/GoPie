package passes

import (
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"strings"
	"toolkit/pkg/inst"
	"toolkit/pkg/utils/gofmt"

	"golang.org/x/tools/go/ast/astutil"
)

func RunTestPass(in, out string) error {
	p := TestPass{}
	iCtx, err := inst.NewInstContext(in)
	if err != nil {
		log.Fatalf("Analysis source code failed %v", err)
	}
	p.Before(iCtx)
	iCtx.AstFile = astutil.Apply(iCtx.AstFile, p.GetPreApply(iCtx), p.GetPostApply(iCtx)).(*ast.File)
	p.After(iCtx)

	inst.DumpAstFile(iCtx.FS, iCtx.AstFile, out)
	if gofmt.HasSyntaxError(out) {
		err = ioutil.WriteFile(out, iCtx.OriginalContent, 0777)
		if err != nil {
			log.Panicf("failed to recover file '%s'", out)
		}
		// do_retry(out, out, wp)
	}
	return nil
}

type TestPass struct {
	Pos string
}

var (
	TestNeedInst   = "NEED_TEST_INST"
	TestImportName = "sched"
	TestImportPath = "sched"
)

func (p *TestPass) Before(ctx *inst.InstContext) {
	ctx.SetMetadata(TestNeedInst, false)
}

func (p *TestPass) After(ctx *inst.InstContext) {
	if v, ok := ctx.GetMetadata(TestNeedInst); ok && v.(bool) {
		inst.AddImport(ctx.FS, ctx.AstFile, TestImportName, TestImportPath)
	}
}

func (p *TestPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.FuncDecl:
			name := concrete.Name.Name
			params := concrete.Type.Params.List

			if len(params) != 1 {
				return true
			}

			check_ok := false
			names := params[0].Names
			if len(names) != 1 || names[0].Name != "t" {
				return false
			}

			if v, ok := params[0].Type.(*ast.StarExpr); ok {
				if vv, ok := v.X.(*ast.SelectorExpr); ok {
					if vvv, ok := vv.X.(*ast.Ident); ok {
						if vv.Sel.Name == "T" && vvv.Name == "testing" {
							check_ok = true
						}
					}
				}
			}
			if check_ok && strings.HasPrefix(name, "Test") && !strings.HasSuffix(name, "_1") {
				testname := name
				testfunc := concrete

				var testDecl ast.Decl
				if p.Pos == "inside" {
					testDecl = genTestDeclWithTimeout(testname, testfunc)
				} else {
					testDecl = genTestDeclWithoutTimeout(testname, testfunc)
				}
				iCtx.AstFile.Decls = append(iCtx.AstFile.Decls, testDecl)
				iCtx.SetMetadata(TestNeedInst, true)
			}
		}
		return true
	}
}

func (p *TestPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func genTestDeclWithoutTimeout(name string, fn *ast.FuncDecl) *ast.FuncDecl {
	testname := name + "_1"

	checker := NewDeferExpr("sched", "Leakcheck", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.IDENT,
			Value: "t",
		},
	})

	testbodylst := make([]ast.Stmt, len(fn.Body.List))
	copy(testbodylst, fn.Body.List)

	parseinput := &ast.ExprStmt{NewArgCall("sched", "ParseInput", []ast.Expr{})}

	testbody := []ast.Stmt{parseinput, checker}
	testbody = append(testbody, testbodylst...)
	block := &ast.BlockStmt{List: testbody}

	testdecl := &ast.FuncDecl{
		Name: &ast.Ident{Name: testname},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{Name: "t"},
						},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "testing",
								},
								Sel: &ast.Ident{
									Name: "T",
								},
							},
						},
					},
				},
			},
		},
		Body: block,
	}
	return testdecl
}

func genTestDeclWithTimeout(name string, fn *ast.FuncDecl) *ast.FuncDecl {
	testname := name + "_1"

	decldone := &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{&ast.Ident{Name: "done_xxx"}},
		Rhs: []ast.Expr{NewArgCall("sched", "GetDone", []ast.Expr{})},
	}

	decltimeout := &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{&ast.Ident{Name: "timeout_xxx"}},
		Rhs: []ast.Expr{NewArgCall("sched", "GetTimeout", []ast.Expr{})},
	}

	testgodone := NewDeferExpr("sched", "Done", []ast.Expr{
		&ast.Ident{Name: "done_xxx"},
	})

	checker := NewDeferExpr("sched", "Leakcheck", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.IDENT,
			Value: "t",
		},
	})

	testgobodylst := make([]ast.Stmt, len(fn.Body.List))
	copy(testgobodylst, fn.Body.List)

	testgobodylst = append([]ast.Stmt{testgodone, checker}, fn.Body.List...)
	testgobody := &ast.BlockStmt{List: testgobodylst}

	testgo := &ast.GoStmt{Call: &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{},
				},
			},
			Body: testgobody,
		},
	}}

	parseinput := &ast.ExprStmt{NewArgCall("sched", "ParseInput", []ast.Expr{})}

	testselect := &ast.SelectStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.CommClause{
					Comm: &ast.ExprStmt{
						X: &ast.UnaryExpr{
							Op: token.ARROW,
							X:  &ast.Ident{Name: "timeout_xxx"},
						},
					},
				},
				&ast.CommClause{
					Comm: &ast.ExprStmt{
						X: &ast.UnaryExpr{
							Op: token.ARROW,
							X:  &ast.Ident{Name: "done_xxx"},
						},
					},
				},
			},
		},
	}

	testbody := &ast.BlockStmt{List: []ast.Stmt{
		parseinput,
		decldone,
		decltimeout,
		testgo,
		testselect,
	},
	}

	testdecl := &ast.FuncDecl{
		Name: &ast.Ident{Name: testname},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{Name: "t"},
						},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "testing",
								},
								Sel: &ast.Ident{
									Name: "T",
								},
							},
						},
					},
				},
			},
		},
		Body: testbody,
	}
	return testdecl
}

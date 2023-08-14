package passes

import (
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"log"
	"toolkit/pkg/inst"
	"toolkit/pkg/utils/gofmt"
)

// ChResPass, Channel Record Pass. This pass instrumented at
// following four channel related operations:
// send, recv, make, close

var (
	SelectInstNeed   = "SelectNeedInst"
	SelectImportName = "sched"
	SelectImportPath = "sched"
)

type SelectPass struct {
}

func RunSelectPass(in, out string) error {
	p := SelectPass{}
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

func (p *SelectPass) Before(iCtx *inst.InstContext) {
	iCtx.SetMetadata(SelectInstNeed, false)
}

func (p *SelectPass) After(iCtx *inst.InstContext) {
	need, _ := iCtx.GetMetadata(SelectInstNeed)
	needinst := need.(bool)
	if needinst {
		inst.AddImport(iCtx.FS, iCtx.AstFile, SelectImportName, SelectImportPath)
	}
}

func (p *SelectPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *SelectPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {

		// channel send operation
		case *ast.SelectStmt:
			cases := concrete.Body.List
			for _, x := range cases {
				if _, ok := x.(*ast.CommClause); !ok {
					return true
				}
				comm, _ := x.(*ast.CommClause)
				switch concrete := comm.Comm.(type) {
				case *ast.ExprStmt: // recv
					id := iCtx.GetNewOpId()
					unaryExpr, _ := concrete.X.(*ast.UnaryExpr)
					ch := unaryExpr.X
					Add(concrete.Pos(), id)
					newCall := GenInstCall("InstChAF", ch, id)
					comm.Body = append([]ast.Stmt{newCall}, comm.Body...)
					iCtx.SetMetadata(SelectInstNeed, true)
				case *ast.AssignStmt:
					id := iCtx.GetNewOpId()
					var unaryExpr *ast.UnaryExpr
					for _, rhs := range concrete.Rhs {
						if v, ok := rhs.(*ast.UnaryExpr); ok {
							unaryExpr = v
						}
					}
					if unaryExpr != nil {
						ch := unaryExpr.X
						Add(concrete.Pos(), id)
						newCall := GenInstCall("InstChAF", ch, id)
						comm.Body = append([]ast.Stmt{newCall}, comm.Body...)
						iCtx.SetMetadata(SelectInstNeed, true)
					}
				case *ast.SendStmt: // send
					id := iCtx.GetNewOpId()
					Add(concrete.Pos(), id)
					ch := concrete.Chan
					newCall := GenInstCall("InstChAF", ch, id)
					comm.Body = append([]ast.Stmt{newCall}, comm.Body...)
					iCtx.SetMetadata(SelectInstNeed, true)
				}
			}
		}

		return true
	}
}

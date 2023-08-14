package passes

import (
	"go/ast"
	"go/token"
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
	ChannelNeedInst   = "ChannelNeedInst"
	ChannelImportName = "sched"
	ChannelImportPath = "sched"
)

type ChRecPass struct {
}

func RunChannelPass(in, out string) error {
	p := ChRecPass{}
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

func (p *ChRecPass) Before(iCtx *inst.InstContext) {
	iCtx.SetMetadata(ChannelNeedInst, false)
}

func (p *ChRecPass) After(iCtx *inst.InstContext) {
	need, _ := iCtx.GetMetadata(ChannelNeedInst)
	needinst := need.(bool)
	if needinst {
		inst.AddImport(iCtx.FS, iCtx.AstFile, ChannelImportName, ChannelImportPath)
	}
}

func (p *ChRecPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *ChRecPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {

		// channel send operation
		case *ast.SendStmt:
			id := iCtx.GetNewOpId()
			Add(concrete.Pos(), id)
			ch := concrete.Chan
			before := GenInstCall("InstChBF", ch, id)
			c.InsertBefore(before)
			after := GenInstCall("InstChAF", ch, id)
			c.InsertAfter(after)

			iCtx.SetMetadata(ChannelNeedInst, true)

		// channel recv operation
		case *ast.ExprStmt:
			if unaryExpr, ok := concrete.X.(*ast.UnaryExpr); ok {
				if unaryExpr.Op == token.ARROW { // This is a receive operation
					id := iCtx.GetNewOpId()
					Add(concrete.Pos(), id)
					ch := unaryExpr.X
					before := GenInstCall("InstChBF", ch, id)
					c.InsertBefore(before)

					after := GenInstCall("InstChAF", ch, id)
					c.InsertAfter(after)
					iCtx.SetMetadata(ChannelNeedInst, true)
				}
			} else if callExpr, ok := concrete.X.(*ast.CallExpr); ok { // like `close(ch)` or `mu.Lock()`
				if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok { // like `close(ch)`
					// channel close operation
					if funcIdent.Name == "close" {
						id := iCtx.GetNewOpId()
						Add(concrete.Pos(), id)
						args := callExpr.Args
						if len(args) == 1 {
							if ch, ok := args[0].(*ast.Ident); ok {

								before := GenInstCall("InstChBF", ch, id)
								c.InsertBefore(before)

								after := GenInstCall("InstChAF", ch, id)
								c.InsertAfter(after)

								iCtx.SetMetadata(ChannelNeedInst, true)
							}
						}
					}
				}
			}
		case *ast.DeferStmt:
			callExpr := concrete.Call
			if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok { // like `close(ch)`
				// channel close operation
				if funcIdent.Name == "close" {
					id := iCtx.GetNewOpId()
					Add(concrete.Pos(), id)
					args := callExpr.Args
					if len(args) == 1 {
						if ch, ok := args[0].(*ast.Ident); ok {
							before := GenInstCall("InstChBF", ch, id)
							after := GenInstCall("InstChAF", ch, id)

							body := &ast.BlockStmt{List: []ast.Stmt{
								before,
								NewArgCallExpr("", "close", callExpr.Args),
								after,
							}}

							deferStmt := &ast.DeferStmt{
								Call: &ast.CallExpr{
									Fun: &ast.FuncLit{
										Type: &ast.FuncType{Params: &ast.FieldList{List: nil}},
										Body: body,
									},
									Args: []ast.Expr{},
								},
							}

							c.Replace(deferStmt)
							iCtx.SetMetadata(ChannelNeedInst, true)
						}
					}
				}
			}
			return false
		}
		return true
	}
}

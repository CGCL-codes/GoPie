package passes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"toolkit/pkg/inst"
)

var (
	LockNeedInst   = "LockNeedInst"
	LockImportName = "sched"
	LockImportPath = "sched"
)

type LockPass struct {
}

func (p *LockPass) Before(iCtx *inst.InstContext) {
	iCtx.SetMetadata(LockNeedInst, false)
}

func (p *LockPass) After(iCtx *inst.InstContext) {
	need, _ := iCtx.GetMetadata(LockNeedInst)
	needinst := need.(bool)
	if needinst {
		inst.AddImport(iCtx.FS, iCtx.AstFile, LockImportName, LockImportPath)
	}
}

func (p *LockPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *LockPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.ExprStmt:
			if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
				if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
					if SelectorCallerHasTypes(iCtx, selectorExpr, true, "sync.Mutex", "*sync.Mutex", "sync.RWMutex", "*sync.RWMutex") {
						var matched bool = true
						switch selectorExpr.Sel.Name {
						case "Lock":
						case "RLock":
						case "RUnlock":
						case "Unlock":
						default:
							matched = false
						}

						if matched {
							id := iCtx.GetNewOpId()
							Add(concrete.Pos(), id)
							mu := selectorExpr.X
							p_mu := &ast.UnaryExpr{
								Op: token.AND,
								X:  mu,
							}
							before := GenInstCall("InstMutexBF", p_mu, id)
							c.InsertBefore(before)
							after := GenInstCall("InstMutexAF", p_mu, id)
							c.InsertAfter(after)
							iCtx.SetMetadata(LockNeedInst, true)
						}
					}
				}
			}
		case *ast.DeferStmt:
			callExpr := concrete.Call
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				if SelectorCallerHasTypes(iCtx, selectorExpr, true, "sync.Mutex", "*sync.Mutex", "sync.RWMutex", "*sync.RWMutex") {
					var matched bool = true
					switch selectorExpr.Sel.Name {
					case "Lock":
					case "RLock":
					case "RUnlock":
					case "Unlock":
					default:
						matched = false
					}

					if matched {
						id := iCtx.GetNewOpId()
						Add(concrete.Pos(), id)

						mu := selectorExpr.X
						p_mu := &ast.UnaryExpr{
							Op: token.AND,
							X:  mu,
						}
						before := GenInstCall("InstMutexBF", p_mu, id)
						after := GenInstCall("InstMutexAF", p_mu, id)

						body := &ast.BlockStmt{List: []ast.Stmt{
							before,
							&ast.ExprStmt{callExpr},
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
						iCtx.SetMetadata(LockNeedInst, true)
					}
				}
			}
			return false
		}
		return true
	}
}

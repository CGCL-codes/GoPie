package passes

import (
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"toolkit/pkg/inst"
)

func RunFLWrapperPass(in, out string, wp *WrapperPass) error {
	p := wp
	iCtx, err := inst.NewInstContext(in)
	if err != nil {
		log.Fatalf("Analysis source code failed %v", err)
	}
	p.Before(iCtx)
	iCtx.AstFile = astutil.Apply(iCtx.AstFile, p.GetPreApply(iCtx), p.GetPostApply(iCtx)).(*ast.File)
	p.After(iCtx)

	inst.DumpAstFile(iCtx.FS, iCtx.AstFile, out)
	return nil
}

type Import struct {
	Name string
	Need string
	Path string
}

type WrapperPass struct {
	before *ast.Stmt
	after  *ast.DeferStmt
	dowrap func(node ast.Node) bool
	Import
}

func NewWrapperPass(before *ast.Stmt, after *ast.DeferStmt, dowrap func(node ast.Node) bool, imp Import) *WrapperPass {
	wp := &WrapperPass{
		before: before,
		after:  after,
		dowrap: dowrap,
	}
	wp.Name = imp.Name
	wp.Need = imp.Need
	wp.Path = imp.Path
	return wp
}

func (p *WrapperPass) Before(ctx *inst.InstContext) {
	if p.Need == "" {
		return
	}
	ctx.SetMetadata(p.Need, false)
}

func (p *WrapperPass) After(ctx *inst.InstContext) {
	if p.Need == "" {
		return
	}
	if v, ok := ctx.GetMetadata(p.Need); ok && v.(bool) {
		inst.AddImport(ctx.FS, ctx.AstFile, p.Name, p.Path)
	}
}

func (p *WrapperPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.FuncDecl:
			if p.dowrap(concrete) {
				if p.after != nil {
					concrete.Body.List = append([]ast.Stmt{p.after}, concrete.Body.List...)
					iCtx.SetMetadata(p.Need, true)
				}
				if p.before != nil {
					concrete.Body.List = append([]ast.Stmt{*p.before}, concrete.Body.List...)
					iCtx.SetMetadata(p.Need, true)
				}
			}
		}

		return true
	}
}

func (p *WrapperPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

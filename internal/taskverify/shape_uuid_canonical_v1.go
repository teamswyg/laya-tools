package taskverify

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// uuidCanonicalSupportedSource deliberately supports a small append-only shape.
// It is not a general Go effect checker: other implementations are unknown,
// including correct implementations that alter legacy code or add helpers.
func uuidCanonicalSupportedSource(source []byte) bool {
	// The public original is not gofmt-clean under 1.27.1. Both exact trusted
	// prefixes preserve its legacy semantics/header; only the fixed formatter's
	// whitespace normalization is allowed, not arbitrary legacy rewriting.
	prefixBytes := 0
	switch {
	case len(source) >= 10254 && digest(source[:10254]) == "00d93a3f65d8b0063266e63ff0e16ac543e30a911fd35119bcf1a66437161360":
		prefixBytes = 10254
	case len(source) >= 10251 && digest(source[:10251]) == "b19f8aad57fbe17219758f0142742fb3e1879d26506f7fce99169b3ec770c7e8":
		prefixBytes = 10251
	default:
		return false
	}
	suffix := source[prefixBytes:]
	if strings.TrimSpace(string(suffix)) == "" {
		return true
	} // assessed baseline: API is absent
	file, err := parser.ParseFile(token.NewFileSet(), "canonical-addition.go", append([]byte("package uuid\n"), suffix...), parser.ParseComments)
	if err != nil || len(file.Decls) != 1 || len(file.Imports) != 0 {
		return false
	}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(c.Text, "//")), "go:") {
				return false
			}
		}
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || fn.Name.Name != "ParseCanonical" || fn.Recv != nil || fn.Body == nil || fn.Type.TypeParams != nil {
		return false
	}
	params, results := fn.Type.Params.List, fn.Type.Results
	if len(params) != 1 || len(params[0].Names) != 1 || params[0].Names[0].Name != "s" || !uuidCanonicalIdent(params[0].Type, "string") || results == nil || len(results.List) != 2 {
		return false
	}
	if len(results.List[0].Names) != 0 || len(results.List[1].Names) != 0 || !uuidCanonicalIdent(results.List[0].Type, "UUID") || !uuidCanonicalIdent(results.List[1].Type, "error") {
		return false
	}
	locals := map[*ast.Object]bool{params[0].Names[0].Obj: true}
	if params[0].Names[0].Obj == nil {
		return false
	}
	good := true
	addLocal := func(id *ast.Ident) {
		name := id.Name
		if name == "_" {
			return
		}
		if uuidCanonicalReserved(name) || id.Obj == nil {
			good = false
			return
		}
		locals[id.Obj] = true
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, x := range n.Lhs {
					id, ok := x.(*ast.Ident)
					if !ok {
						good = false
					} else {
						addLocal(id)
					}
				}
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				for _, x := range []ast.Expr{n.Key, n.Value} {
					if x == nil {
						continue
					}
					id, ok := x.(*ast.Ident)
					if !ok {
						good = false
					} else {
						addLocal(id)
					}
				}
			}
		case *ast.ValueSpec:
			for _, id := range n.Names {
				addLocal(id)
			}
		}
		return good
	})
	if !good {
		return false
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if !good {
			return false
		}
		switch n := n.(type) {
		case *ast.FuncLit, *ast.FuncType, *ast.GoStmt, *ast.DeferStmt, *ast.SendStmt, *ast.ChanType, *ast.TypeSpec, *ast.StarExpr, *ast.TypeAssertExpr, *ast.InterfaceType:
			good = false
		case *ast.Ident:
			if !locals[n.Obj] && !(n.Obj == nil && uuidCanonicalReserved(n.Name)) {
				good = false
			}
		case *ast.UnaryExpr:
			if n.Op == token.AND || n.Op == token.ARROW {
				good = false
			}
		case *ast.SliceExpr:
			if !uuidCanonicalLocalTarget(n.X, locals) {
				good = false
			} // no alias of a global array
		case *ast.SelectorExpr:
			id, ok := n.X.(*ast.Ident)
			if !ok || !uuidCanonicalPureSelector(id, n.Sel.Name, locals) {
				good = false
			}
			return false // field name is not an unrelated identifier
		case *ast.CallExpr:
			switch f := n.Fun.(type) {
			case *ast.Ident:
				if !uuidCanonicalPureCall(f.Name) {
					good = false
				}
				if (f.Name == "copy" || f.Name == "append") && (len(n.Args) == 0 || !uuidCanonicalLocalTarget(n.Args[0], locals)) {
					good = false
				}
			case *ast.SelectorExpr:
				id, ok := f.X.(*ast.Ident)
				if !ok || !uuidCanonicalPureSelector(id, f.Sel.Name, locals) {
					good = false
				}
				if ok && id.Name == "hex" && f.Sel.Name == "Decode" && (len(n.Args) == 0 || !uuidCanonicalLocalTarget(n.Args[0], locals)) {
					good = false
				}
			default:
				good = false
			}
		case *ast.AssignStmt:
			for _, x := range n.Lhs {
				if !uuidCanonicalLocalTarget(x, locals) {
					good = false
				}
			}
		case *ast.RangeStmt:
			for _, x := range []ast.Expr{n.Key, n.Value} {
				if x != nil && !uuidCanonicalLocalTarget(x, locals) {
					good = false
				}
			}
		case *ast.IncDecStmt:
			if !uuidCanonicalLocalTarget(n.X, locals) {
				good = false
			}
		}
		return good
	})
	return good
}

func uuidCanonicalIdent(x ast.Expr, name string) bool {
	id, ok := x.(*ast.Ident)
	return ok && id.Name == name
}
func uuidCanonicalReserved(name string) bool {
	switch name {
	case "_", "nil", "true", "false", "UUID", "error", "byte", "rune", "int", "uint", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "string", "bool",
		"Nil", "Max", "ErrInvalidUUIDFormat", "RFC4122", "Parse", "xtob", "len", "cap", "copy", "append", "make", "strings", "bytes", "hex", "fmt", "errors":
		return true
	}
	return false
}
func uuidCanonicalPureCall(name string) bool {
	switch name {
	case "Parse", "xtob", "len", "cap", "copy", "append", "make", "UUID", "byte", "rune", "int", "uint", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "string", "bool":
		return true
	}
	return false
}
func uuidCanonicalPureSelector(id *ast.Ident, name string, locals map[*ast.Object]bool) bool {
	base := id.Name
	if locals[id.Obj] || (id.Obj == nil && (base == "Nil" || base == "Max")) {
		return name == "String" || name == "Variant"
	}
	switch base {
	case "strings":
		switch name {
		case "ToLower", "ToUpper", "TrimSpace", "TrimPrefix", "TrimSuffix", "Trim", "ReplaceAll", "Replace", "EqualFold", "HasPrefix", "HasSuffix", "Index", "IndexByte", "Contains", "Count", "Repeat":
			return true
		}
	case "hex":
		switch name {
		case "Decode", "DecodeString", "EncodeToString", "DecodedLen":
			return true
		}
	case "bytes":
		switch name {
		case "Equal", "Index", "IndexByte", "Contains", "Count", "HasPrefix", "HasSuffix":
			return true
		}
	case "fmt":
		return name == "Errorf" || name == "Sprintf"
	case "errors":
		return name == "New"
	}
	return false
}
func uuidCanonicalLocalTarget(x ast.Expr, locals map[*ast.Object]bool) bool {
	switch x := x.(type) {
	case *ast.Ident:
		return x.Name == "_" || (x.Obj != nil && locals[x.Obj])
	case *ast.IndexExpr:
		return uuidCanonicalLocalTarget(x.X, locals)
	case *ast.SliceExpr:
		return uuidCanonicalLocalTarget(x.X, locals)
	case *ast.ParenExpr:
		return uuidCanonicalLocalTarget(x.X, locals)
	}
	return false
}

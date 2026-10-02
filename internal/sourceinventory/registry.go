package sourceinventory

import (
	"go/ast"
	"go/token"
	"strconv"
)

func stringLiteral(e ast.Expr) (string, error) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", ErrRegistry
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil || s == "" {
		return "", ErrRegistry
	}
	return s, nil
}

func boolLiteral(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && (id.Name == "true" || id.Name == "false")
}

func topValue(f *ast.File, name string) (*ast.CompositeLit, error) {
	var result *ast.CompositeLit
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			v, ok := spec.(*ast.ValueSpec)
			if !ok {
				return nil, ErrRegistry
			}
			for _, n := range v.Names {
				if n.Name != name {
					continue
				}
				if result != nil || len(v.Names) != 1 || len(v.Values) != 1 {
					return nil, ErrRegistry
				}
				var ok bool
				result, ok = v.Values[0].(*ast.CompositeLit)
				if !ok {
					return nil, ErrRegistry
				}
			}
		}
	}
	if result == nil {
		return nil, ErrRegistry
	}
	return result, nil
}

func topFunction(f *ast.File, name string) (*ast.FuncDecl, error) {
	var result *ast.FuncDecl
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != name {
			continue
		}
		if result != nil || fn.Body == nil {
			return nil, ErrRegistry
		}
		result = fn
	}
	if result == nil {
		return nil, ErrRegistry
	}
	return result, nil
}

func recordLiteral(lit *ast.CompositeLit, cohort string, run ast.Expr) (registryEntry, error) {
	if len(lit.Elts) != 5 || !boolLiteral(lit.Elts[4]) {
		return registryEntry{}, ErrRegistry
	}
	var s [4]string
	for i := 0; i < 4; i++ {
		text, e := stringLiteral(lit.Elts[i])
		if e != nil {
			return registryEntry{}, e
		}
		s[i] = text
	}
	if run != nil {
		id, ok := run.(*ast.Ident)
		if !ok || id.Name != s[3] {
			return registryEntry{}, ErrRegistry
		}
	}
	return registryEntry{cohort, s[0], s[1], s[2], s[3]}, nil
}

func readRegistry(files [4]*ast.File) ([]registryEntry, error) {
	var out []registryEntry
	legacy, e := topValue(files[0], "sourceCatalog")
	if e != nil {
		return nil, e
	}
	array, ok := legacy.Type.(*ast.ArrayType)
	if !ok || array.Len != nil {
		return nil, ErrRegistry
	}
	element, ok := array.Elt.(*ast.Ident)
	if !ok || element.Name != "sourceDef" {
		return nil, ErrRegistry
	}
	for _, expr := range legacy.Elts {
		row, ok := expr.(*ast.CompositeLit)
		if !ok || row.Type != nil || len(row.Elts) != 6 {
			return nil, ErrRegistry
		}
		copy := *row
		copy.Elts = append(append([]ast.Expr{}, row.Elts[:4]...), row.Elts[5])
		r, e := recordLiteral(&copy, "legacy", row.Elts[4])
		if e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	for index, name := range []string{"atomicSources", "identitySources", "snapshotSources"} {
		lit, e := topValue(files[2], name)
		if e != nil {
			return nil, e
		}
		a, ok := lit.Type.(*ast.ArrayType)
		if !ok || a.Len != nil {
			return nil, ErrRegistry
		}
		st, ok := a.Elt.(*ast.StructType)
		if !ok || len(st.Fields.List) != 2 {
			return nil, ErrRegistry
		}
		first, second := st.Fields.List[0], st.Fields.List[1]
		if len(first.Names) != 1 || first.Names[0].Name != "spec" || len(second.Names) != 1 || second.Names[0].Name != "run" {
			return nil, ErrRegistry
		}
		t, ok := first.Type.(*ast.Ident)
		if !ok || t.Name != "sourceSpec" {
			return nil, ErrRegistry
		}
		runType, ok := second.Type.(*ast.Ident)
		if !ok || runType.Name != [3]string{"atomicFn", "identityFn", "snapshotFn"}[index] {
			return nil, ErrRegistry
		}
		for _, expr := range lit.Elts {
			row, ok := expr.(*ast.CompositeLit)
			if !ok || row.Type != nil || len(row.Elts) != 2 {
				return nil, ErrRegistry
			}
			record, ok := row.Elts[0].(*ast.CompositeLit)
			if !ok {
				return nil, ErrRegistry
			}
			name, ok := record.Type.(*ast.Ident)
			if !ok || name.Name != "sourceSpec" {
				return nil, ErrRegistry
			}
			r, e := recordLiteral(record, "typed", row.Elts[1])
			if e != nil {
				return nil, e
			}
			out = append(out, r)
		}
	}
	flow, e := topFunction(files[3], "sourcesFlow")
	if e != nil || len(flow.Body.List) != 1 {
		return nil, ErrRegistry
	}
	ret, ok := flow.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return nil, ErrRegistry
	}
	lit, ok := ret.Results[0].(*ast.CompositeLit)
	if !ok {
		return nil, ErrRegistry
	}
	a, ok := lit.Type.(*ast.ArrayType)
	if !ok || a.Len != nil {
		return nil, ErrRegistry
	}
	t, ok := a.Elt.(*ast.Ident)
	if !ok || t.Name != "sourceSpec" {
		return nil, ErrRegistry
	}
	var flowEntries []registryEntry
	for _, expr := range lit.Elts {
		row, ok := expr.(*ast.CompositeLit)
		if !ok || row.Type != nil {
			return nil, ErrRegistry
		}
		r, e := recordLiteral(row, "typed", nil)
		if e != nil {
			return nil, e
		}
		flowEntries = append(flowEntries, r)
	}
	if e := checkFlowDispatch(files[3], flowEntries); e != nil {
		return nil, e
	}
	out = append(out, flowEntries...)
	for i, r := range out {
		for j := 0; j < i; j++ {
			if r.cohort == out[j].cohort && r.id == out[j].id {
				return nil, ErrRegistry
			}
		}
	}
	return out, nil
}

func checkFlowDispatch(file *ast.File, entries []registryEntry) error {
	fn, e := topFunction(file, "checkFlow")
	if e != nil {
		return e
	}
	if !dispatchSignature(fn.Type) {
		return ErrRegistry
	}
	var sw *ast.SwitchStmt
	for _, stmt := range fn.Body.List {
		switch s := stmt.(type) {
		case *ast.SwitchStmt:
			if sw != nil {
				return ErrRegistry
			}
			sw = s
		case *ast.DeferStmt:
			// The only supported prelude is the historical observer's panic
			// recorder. No declaration, assignment or arbitrary call can shadow
			// a candidate identifier before the direct dispatch switch.
			if !panicRecorder(s) {
				return ErrRegistry
			}
		default:
			return ErrRegistry
		}
	}
	if sw == nil || sw.Init != nil {
		return ErrRegistry
	}
	tag, ok := sw.Tag.(*ast.Ident)
	if !ok || tag.Name != "id" {
		return ErrRegistry
	}
	var seen []string
	defaults := 0
	for _, stmt := range sw.Body.List {
		c, ok := stmt.(*ast.CaseClause)
		if !ok {
			return ErrRegistry
		}
		if len(c.List) == 0 {
			defaults++
			continue
		}
		if len(c.List) != 1 || len(c.Body) != 1 {
			return ErrRegistry
		}
		id, e := stringLiteral(c.List[0])
		if e != nil {
			return e
		}
		for _, old := range seen {
			if old == id {
				return ErrRegistry
			}
		}
		seen = append(seen, id)
		r, ok := c.Body[0].(*ast.ReturnStmt)
		if !ok || len(r.Results) != 1 {
			return ErrRegistry
		}
		call, ok := r.Results[0].(*ast.CallExpr)
		if !ok || len(call.Args) != 2 || call.Ellipsis.IsValid() {
			return ErrRegistry
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != "checkLifecycle" && callee.Name != "checkQuoted" && callee.Name != "checkGraph" {
			return ErrRegistry
		}
		root, ok := call.Args[0].(*ast.Ident)
		if !ok {
			return ErrRegistry
		}
		found := false
		for _, entry := range entries {
			if entry.id == id {
				found = entry.function == root.Name
			}
		}
		if !found {
			return ErrRegistry
		}
	}
	if len(seen) != len(entries) || defaults != 1 {
		return ErrRegistry
	}
	return nil
}

func dispatchSignature(fn *ast.FuncType) bool {
	if fn.TypeParams != nil || fn.Params == nil || len(fn.Params.List) != 1 || fn.Results == nil || len(fn.Results.List) != 1 {
		return false
	}
	p := fn.Params.List[0]
	if len(p.Names) != 1 || p.Names[0].Name != "id" {
		return false
	}
	name, ok := p.Type.(*ast.Ident)
	if !ok || name.Name != "string" {
		return false
	}
	r := fn.Results.List[0]
	if len(r.Names) > 1 || len(r.Names) == 1 && r.Names[0].Name != "out" {
		return false
	}
	result, ok := r.Type.(*ast.Ident)
	return ok && (result.Name == "controlCheck" && len(r.Names) == 1 && r.Names[0].Name == "out" || result.Name == "int" && len(r.Names) == 0)
}

func panicRecorder(stmt *ast.DeferStmt) bool {
	call := stmt.Call
	if len(call.Args) != 0 || call.Ellipsis.IsValid() {
		return false
	}
	fn, ok := call.Fun.(*ast.FuncLit)
	if !ok || fn.Type.Params == nil || len(fn.Type.Params.List) != 0 || fn.Type.Results != nil || len(fn.Body.List) != 1 {
		return false
	}
	cond, ok := fn.Body.List[0].(*ast.IfStmt)
	if !ok || cond.Init != nil || cond.Else != nil || len(cond.Body.List) != 1 {
		return false
	}
	compare, ok := cond.Cond.(*ast.BinaryExpr)
	if !ok || compare.Op != token.NEQ {
		return false
	}
	check, ok := compare.X.(*ast.CallExpr)
	if !ok || len(check.Args) != 0 || check.Ellipsis.IsValid() {
		return false
	}
	name, ok := check.Fun.(*ast.Ident)
	if !ok || name.Name != "recover" {
		return false
	}
	nilValue, ok := compare.Y.(*ast.Ident)
	if !ok || nilValue.Name != "nil" {
		return false
	}
	assign, ok := cond.Body.List[0].(*ast.AssignStmt)
	if !ok || assign.Tok != token.ASSIGN || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return false
	}
	field, ok := assign.Lhs[0].(*ast.SelectorExpr)
	if !ok || field.Sel.Name != "Unknown" {
		return false
	}
	out, ok := field.X.(*ast.Ident)
	if !ok || out.Name != "out" {
		return false
	}
	value, ok := assign.Rhs[0].(*ast.Ident)
	return ok && value.Name == "true"
}

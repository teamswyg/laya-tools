// Package sourceinventory rebinds stored metadata to declarations. It never
// evaluates source, discovers a semantic closure or approves caption fidelity.
package sourceinventory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"slices"
	"strconv"
	"strings"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInput         Error = "inventory_input_invalid"
	ErrFilePin       Error = "inventory_file_pin_mismatch"
	ErrParse         Error = "inventory_source_parse_failed"
	ErrShape         Error = "inventory_source_shape_unsupported"
	ErrRegistry      Error = "inventory_registry_mismatch"
	ErrDeclaration   Error = "inventory_declaration_missing_or_ambiguous"
	ErrHistoricalPin Error = "inventory_historical_pin_mismatch"
	ErrRecipe        Error = "inventory_recipe_unsupported"
)

func fixedPaths() [4]string {
	return [4]string{"internal/behaviorprobe/data.go", "internal/typedbehavior/spec.go", "internal/typedbehavior/state.go", "internal/typedbehavior/flow.go"}
}

type File struct {
	Path           string
	Raw            []byte
	ExpectedSHA256 string
}

type Component struct {
	ID                       string
	Kind                     string
	SHA256                   string
	NormalizedBehaviorSHA256 string
}

type HistoricalRoot struct {
	Cohort               string
	ID                   string
	Prototype            string
	Core                 string
	Function             string // Empty only for legacy metadata, which lacks this field.
	CodeSHA256           string
	NormalizedCodeSHA256 string
	BundleSHA256         string
	Components           []Component
	StandardImports      []string
}

type Input struct {
	Files     [4]File
	Roots     []HistoricalRoot
	GoVersion string
}

type FileReference struct {
	Path   string
	Bytes  int
	SHA256 string
}

type DeclarationReference struct {
	ID               string
	Path             string
	Kind             string
	Symbol           string
	StartByte        int
	EndByte          int
	StartLine        int
	EndLine          int
	RawSHA256        string
	FormattedSHA256  string
	NormalizedSHA256 string
}

type RootReference struct {
	ID                     string
	Cohort                 string
	Root                   DeclarationReference
	Components             []DeclarationReference
	HistoricalBundleSHA256 string
	MetadataPinsMatched    bool
	ContentReview          string
}

type Counters struct {
	FileHashAttempts       int
	FileHashesCompleted    int
	FilePinsMatched        int
	ParseAttempts          int
	ParsesCompleted        int
	ParsesSucceeded        int
	RegistryEntries        int
	RootAttempts           int
	RootsCompleted         int
	ComponentAttempts      int
	ComponentsCompleted    int
	RawSpanHashAttempts    int
	RawSpanHashesCompleted int
	FormatCallsAttempted   int
	FormatCallsCompleted   int
	ScanCallsAttempted     int
	ScanCallsCompleted     int
	BundleCallsAttempted   int
	BundleCallsCompleted   int
}

type Report struct {
	Schema                    string
	State                     string
	Files                     []FileReference
	Roots                     []RootReference
	Counters                  Counters
	ObjectBindingPerformed    bool
	ClosureDiscoveryPerformed bool
	ContentReview             string
	ReusePolicy               string
}

type declaration struct {
	path, kind, symbol string
	node               ast.Node
	file               int
}

type registryEntry struct {
	cohort, id, prototype, core, function string
}

func digest(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func validSHA(s string) bool {
	if len(s) != 64 || strings.ToLower(s) != s {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// Rebind accepts raw bytes with caller-frozen file digests and historical pins.
// Root cardinality is the caller's sealed-plan responsibility. Synthetic tests
// may use smaller registries. No importer, Go subprocess or callback is used.
func Rebind(in Input) (out Report, err error) {
	out = Report{Schema: "riido-source-declaration-reference-draft-v1", State: "incomplete", ContentReview: "pending", ReusePolicy: "no_cache_each_relation_v1"}
	if in.GoVersion != "go1.27.1" || len(in.Roots) == 0 || len(in.Roots) > 120 {
		return out, ErrInput
	}
	paths := fixedPaths()
	fs := token.NewFileSet()
	var parsed [4]*ast.File
	var declarations []declaration
	var imports []string
	for i, file := range in.Files {
		if file.Path != paths[i] || len(file.Raw) == 0 || len(file.Raw) > 1<<20 || !validSHA(file.ExpectedSHA256) {
			return out, ErrInput
		}
		out.Counters.FileHashAttempts++
		actualFileSHA := digest(file.Raw)
		out.Counters.FileHashesCompleted++
		if actualFileSHA != file.ExpectedSHA256 {
			return out, ErrFilePin
		}
		out.Counters.FilePinsMatched++
		out.Files = append(out.Files, FileReference{file.Path, len(file.Raw), file.ExpectedSHA256})
		out.Counters.ParseAttempts++
		f, e := parser.ParseFile(fs, file.Path, file.Raw, 0)
		out.Counters.ParsesCompleted++
		if e != nil {
			return out, ErrParse
		}
		out.Counters.ParsesSucceeded++
		parsed[i] = f
		wantPackage := "typedbehavior"
		if i == 0 {
			wantPackage = "behaviorprobe"
		}
		if f.Name.Name != wantPackage {
			return out, ErrShape
		}
		if i > 0 {
			for _, spec := range f.Imports {
				p, e := strconv.Unquote(spec.Path.Value)
				if e != nil || !slices.Contains([]string{"embed", "errors", "fmt", "strings", "slices", "encoding/json", "encoding/hex", "crypto/sha256"}, p) {
					return out, ErrRecipe
				}
				if spec.Name != nil && spec.Name.Name == "." {
					return out, ErrShape
				}
				if !slices.Contains(imports, p) {
					imports = append(imports, p)
				}
			}
		}
		decls, e := indexDeclarations(f, file.Path, i)
		if e != nil {
			return out, e
		}
		declarations = append(declarations, decls...)
	}
	slices.Sort(imports)
	registry, e := readRegistry(parsed)
	if e != nil {
		return out, e
	}
	out.Counters.RegistryEntries = len(registry)
	if len(registry) != len(in.Roots) {
		return out, ErrRegistry
	}
	var seen []string
	var historicalComponents []Component
	for _, expected := range in.Roots {
		out.Counters.RootAttempts++
		key := expected.Cohort + ":" + expected.ID
		if expected.ID == "" || slices.Contains(seen, key) || !validSHA(expected.CodeSHA256) || !validSHA(expected.NormalizedCodeSHA256) {
			return out, ErrInput
		}
		seen = append(seen, key)
		matches := 0
		var entry registryEntry
		for _, r := range registry {
			if r.cohort == expected.Cohort && r.id == expected.ID {
				entry = r
				matches++
			}
		}
		if matches != 1 || entry.prototype != expected.Prototype || entry.core != expected.Core || expected.Function != "" && entry.function != expected.Function {
			return out, ErrRegistry
		}
		file := 0
		if expected.Cohort == "typed" {
			if expected.Function == "" {
				return out, ErrInput
			}
			file = -1 // Typed root may be declared in state.go or flow.go.
		} else if expected.Cohort != "legacy" {
			return out, ErrInput
		}
		d, e := lookup(declarations, file, "function", entry.function)
		if e != nil {
			return out, e
		}
		ref, e := reference(fs, in.Files, d, "", true, &out.Counters)
		if e != nil {
			return out, e
		}
		if ref.FormattedSHA256 != expected.CodeSHA256 || ref.NormalizedSHA256 != expected.NormalizedCodeSHA256 {
			return out, ErrHistoricalPin
		}
		r := RootReference{ID: expected.ID, Cohort: expected.Cohort, Root: ref, ContentReview: "pending"}
		if expected.Cohort == "legacy" {
			if expected.BundleSHA256 != "" || len(expected.Components) != 0 || len(expected.StandardImports) != 0 {
				return out, ErrInput
			}
		} else {
			if !validSHA(expected.BundleSHA256) || len(expected.Components) == 0 || len(expected.Components) > 256 || !slices.Equal(expected.StandardImports, imports) {
				return out, ErrHistoricalPin
			}
			var componentIDs []string
			rootFound := false
			for _, c := range expected.Components {
				out.Counters.ComponentAttempts++
				if slices.Contains(componentIDs, c.ID) || !validSHA(c.SHA256) || len(componentIDs) > 0 && strings.Compare(componentIDs[len(componentIDs)-1], c.ID) >= 0 {
					return out, ErrInput
				}
				componentIDs = append(componentIDs, c.ID)
				previous := slices.IndexFunc(historicalComponents, func(old Component) bool { return old.ID == c.ID })
				if previous >= 0 && historicalComponents[previous] != c {
					return out, ErrInput
				}
				if previous < 0 {
					historicalComponents = append(historicalComponents, c)
				}
				path, kind, symbol, e := componentIdentity(c.ID)
				if e != nil || kind != c.Kind {
					return out, ErrInput
				}
				file := -1
				for i, p := range paths {
					if p == path {
						file = i
					}
				}
				if file < 1 {
					return out, ErrInput
				}
				decl, e := lookup(declarations, file, kind, symbol)
				if e != nil {
					return out, e
				}
				behavior := kind == "function" || kind == "method"
				component, e := reference(fs, in.Files, decl, c.ID, behavior, &out.Counters)
				if e != nil {
					return out, e
				}
				if c.SHA256 != component.FormattedSHA256 || c.NormalizedBehaviorSHA256 != component.NormalizedSHA256 {
					return out, ErrHistoricalPin
				}
				if decl.file == d.file && kind == "function" && symbol == entry.function {
					rootFound = true
				}
				r.Components = append(r.Components, component)
				out.Counters.ComponentsCompleted++
			}
			if !rootFound {
				return out, ErrHistoricalPin
			}
			out.Counters.BundleCallsAttempted++
			var b bytes.Buffer
			for _, c := range r.Components {
				b.WriteString(c.ID)
				b.WriteByte(0)
				b.WriteString(c.FormattedSHA256)
				b.WriteByte(0)
			}
			for _, p := range imports {
				b.WriteString("std:" + p)
				b.WriteByte(0)
			}
			b.WriteString("go1.27.1;standard-library-infrastructure-v1")
			r.HistoricalBundleSHA256 = digest(b.Bytes())
			out.Counters.BundleCallsCompleted++
			if r.HistoricalBundleSHA256 != expected.BundleSHA256 {
				return out, ErrHistoricalPin
			}
		}
		r.MetadataPinsMatched = true
		out.Roots = append(out.Roots, r)
		out.Counters.RootsCompleted++
	}
	out.State = "metadata_rebound_content_review_pending"
	return out, nil
}

func indexDeclarations(f *ast.File, path string, file int) ([]declaration, error) {
	var out []declaration
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			kind, symbol := "function", d.Name.Name
			if symbol == "init" {
				return nil, ErrShape
			}
			if d.Recv != nil {
				if len(d.Recv.List) != 1 || len(d.Recv.List[0].Names) > 1 {
					return nil, ErrShape
				}
				receiver := d.Recv.List[0].Type
				prefix := ""
				if star, ok := receiver.(*ast.StarExpr); ok {
					prefix = "*"
					receiver = star.X
				}
				name, ok := receiver.(*ast.Ident)
				if !ok {
					return nil, ErrShape
				}
				kind, symbol = "method", prefix+name.Name+"."+symbol
			}
			out = append(out, declaration{path, kind, symbol, d, file})
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					out = append(out, declaration{path, "type", s.Name.Name, s, file})
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name != "_" {
							out = append(out, declaration{path, "value", n.Name, d, file})
						}
					}
				}
			}
		}
	}
	return out, nil
}

func lookup(decls []declaration, file int, kind, symbol string) (declaration, error) {
	count := 0
	var found declaration
	for _, d := range decls {
		if (file < 0 && d.file > 0 || d.file == file) && d.kind == kind && d.symbol == symbol {
			found = d
			count++
		}
	}
	if count != 1 {
		return declaration{}, ErrDeclaration
	}
	return found, nil
}

func reference(fs *token.FileSet, files [4]File, d declaration, id string, behavior bool, counts *Counters) (DeclarationReference, error) {
	start, end := fs.PositionFor(d.node.Pos(), false), fs.PositionFor(d.node.End()-1, false)
	endByte := end.Offset + 1
	raw := files[d.file].Raw
	if start.Offset < 0 || start.Offset >= endByte || endByte > len(raw) || start.Line < 1 || end.Line < start.Line {
		return DeclarationReference{}, ErrShape
	}
	if counts == nil {
		return DeclarationReference{}, ErrInput
	}
	counts.RawSpanHashAttempts++
	rawSHA := digest(raw[start.Offset:endByte])
	counts.RawSpanHashesCompleted++
	var formatted bytes.Buffer
	counts.FormatCallsAttempted++
	formatErr := format.Node(&formatted, fs, d.node)
	counts.FormatCallsCompleted++
	if formatErr != nil {
		return DeclarationReference{}, ErrShape
	}
	r := DeclarationReference{ID: id, Path: d.path, Kind: d.kind, Symbol: d.symbol, StartByte: start.Offset, EndByte: endByte, StartLine: start.Line, EndLine: end.Line, RawSHA256: rawSHA, FormattedSHA256: digest(formatted.Bytes())}
	if behavior {
		counts.ScanCallsAttempted++
		sha, e := normalized(formatted.Bytes())
		counts.ScanCallsCompleted++
		if e != nil {
			return DeclarationReference{}, e
		}
		r.NormalizedSHA256 = sha
	}
	return r, nil
}

func normalized(raw []byte) (string, error) {
	var scan scanner.Scanner
	fs := token.NewFileSet()
	scan.Init(fs.AddFile("code.go", -1, len(raw)), raw, nil, 0)
	var out strings.Builder
	for {
		_, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT && !slices.Contains([]string{"append", "len", "make", "int", "bool", "true", "false", "nil"}, lit) {
			lit = "_"
		}
		out.WriteString(strconv.Itoa(int(tok)))
		out.WriteByte(':')
		out.WriteString(lit)
		out.WriteByte(';')
	}
	if scan.ErrorCount != 0 {
		return "", ErrShape
	}
	return digest([]byte(out.String())), nil
}

func componentIdentity(id string) (path, kind, symbol string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 || parts[2] == "" {
		return "", "", "", ErrInput
	}
	path = "internal/" + parts[0]
	paths := fixedPaths()
	if !slices.Contains(paths[1:], path) {
		return "", "", "", ErrInput
	}
	if !slices.Contains([]string{"function", "method", "type", "value"}, parts[1]) {
		return "", "", "", ErrInput
	}
	return path, parts[1], parts[2], nil
}

package typedbehavior

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

type ComponentPin struct {
	ID                       string `json:"id"`
	Kind                     string `json:"kind"`
	SHA256                   string `json:"sha256"`
	NormalizedBehaviorSHA256 string `json:"normalized_behavior_sha256,omitempty"`
}

// Bundle pins include authored types, globals, methods and referenced helpers.
// Infrastructure affects provenance even when it does not connect task groups.
type SourcePin struct {
	ID                   string         `json:"id"`
	Prototype            string         `json:"prototype"`
	Core                 string         `json:"core_template"`
	Function             string         `json:"function"`
	CodeSHA256           string         `json:"code_sha256"`
	NormalizedCodeSHA256 string         `json:"normalized_code_sha256"`
	BundleSHA256         string         `json:"source_bundle_sha256"`
	Components           []ComponentPin `json:"semantic_components"`
	StandardImports      []string       `json:"standard_imports"`
}

type sourceFile struct{ name, text string }

// Bind the compiled offline implementation as well as the embedded candidate
// source. This prevents a stale audit binary from silently using newer files.
//
//go:embed source.go dataset.go fixtures.go audit.go groups.go
var implementationSource embed.FS

func SourceArtifacts() []SourceArtifact {
	artifacts := []SourceArtifact{{"spec.go", sum([]byte(sourceSpecText))}, {"state.go", sum([]byte(stateSourceText))}, {"flow.go", sum([]byte(flowSourceText))}}
	for _, name := range []string{"source.go", "dataset.go", "fixtures.go", "audit.go", "groups.go"} {
		raw, err := implementationSource.ReadFile(name)
		if err != nil {
			panic("embedded_implementation_source_missing")
		}
		artifacts = append(artifacts, SourceArtifact{name, sum(raw)})
	}
	return artifacts
}

type declaration struct {
	object             types.Object
	node               ast.Node
	file, kind, symbol string
}

// These explicit standard imports are infrastructure, not authored task cores.
// No arbitrary import path or authored symbol can be silently excluded.
var allowedSourceImports = [...]string{
	"embed", "errors", "fmt", "strings", "slices",
	"encoding/json", "encoding/hex", "crypto/sha256",
}

func sourceSpecs() []sourceSpec {
	return append(sourcesState(), sourcesFlow()...)
}

func sum(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

// This is the same token normalization used for v1 candidate-copy grouping.
func normalizedDeclarationSHA(raw []byte) string {
	var scanner scanner.Scanner
	fs := token.NewFileSet()
	scanner.Init(fs.AddFile("code.go", -1, len(raw)), raw, nil, 0)
	var out strings.Builder
	for {
		_, tok, lit := scanner.Scan()
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
	return sum([]byte(out.String()))
}

func formatDeclaration(fs *token.FileSet, n ast.Node) ([]byte, error) {
	var b bytes.Buffer
	if format.Node(&b, fs, n) != nil {
		return nil, errors.New("source_format_failed")
	}
	return b.Bytes(), nil
}

func SourcePins() ([]SourcePin, error) {
	if runtime.Version() != "go1.27.1" {
		return nil, errors.New("source_toolchain_not_pinned")
	}
	files := []sourceFile{{"spec.go", sourceSpecText}, {"state.go", stateSourceText}, {"flow.go", flowSourceText}}
	return pinSources(files, sourceSpecs())
}

// go/types distinguishes globals from shadowing locals and binds method calls.
// Its required metadata maps live only in this offline provenance audit.
func pinSources(files []sourceFile, specs []sourceSpec) ([]SourcePin, error) {
	fs := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(files))
	var imports []string
	for _, file := range files {
		f, e := parser.ParseFile(fs, file.name, file.text, 0)
		if e != nil {
			return nil, errors.New("source_parse_failed")
		}
		for _, i := range f.Imports {
			p, e := strconv.Unquote(i.Path.Value)
			if e != nil || !slices.Contains(allowedSourceImports[:], p) {
				return nil, errors.New("source_import_not_allowlisted")
			}
			if !slices.Contains(imports, p) {
				imports = append(imports, p)
			}
		}
		parsed = append(parsed, f)
	}
	info := &types.Info{Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
	standardImporter, e := sourceImporter(fs, imports)
	if e != nil {
		return nil, e
	}
	config := types.Config{Importer: standardImporter, GoVersion: "go1.27"}
	pkg, e := config.Check("github.com/teamswyg/laya-tools/internal/typedbehavior", fs, parsed, info)
	if e != nil {
		return nil, errors.New("source_typecheck_failed")
	}
	var declarations []declaration
	for index, f := range parsed {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name == "init" {
					return nil, errors.New("source_package_init_unsupported")
				}
				object := info.Defs[d.Name]
				if object == nil {
					return nil, errors.New("source_definition_missing")
				}
				symbol := d.Name.Name
				kind := "function"
				if signature, ok := object.Type().(*types.Signature); ok && signature.Recv() != nil {
					kind = "method"
					symbol = types.TypeString(signature.Recv().Type(), func(*types.Package) string { return "" }) + "." + symbol
				}
				declarations = append(declarations, declaration{object, d, files[index].name, kind, symbol})
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						declarations = append(declarations, declaration{info.Defs[s.Name], s, files[index].name, "type", s.Name.Name})
					case *ast.ValueSpec:
						// Pin the declaration block, including inherited const/iota positions.
						// Unrelated authored initializer calls can mutate a referenced global
						// before the candidate executes; that unsupported pattern fails closed.
						badInitializer := false
						for _, value := range s.Values {
							ast.Inspect(value, func(node ast.Node) bool {
								call, ok := node.(*ast.CallExpr)
								if !ok {
									return true
								}
								switch f := call.Fun.(type) {
								case *ast.Ident:
									object := info.Uses[f]
									if _, ok := object.(*types.TypeName); ok {
										return true
									}
									// Builtins such as copy/append can mutate another
									// global through a blank initializer. Supporting
									// initializer effects requires a separate policy;
									// this closed registry needs no builtin calls here.
								case *ast.SelectorExpr:
									object := info.Uses[f.Sel]
									if function, ok := object.(*types.Func); ok && function.Pkg() != nil && function.Pkg().Path() == "errors" && function.Name() == "New" {
										return true
									}
								}
								badInitializer = true
								return true
							})
						}
						if badInitializer {
							return nil, errors.New("source_global_initializer_call_unsupported")
						}
						for _, name := range s.Names {
							if name.Name != "_" {
								declarations = append(declarations, declaration{info.Defs[name], d, files[index].name, "value", name.Name})
							}
						}
					}
				}
			}
		}
	}
	lookup := func(object types.Object) int {
		return slices.IndexFunc(declarations, func(d declaration) bool { return d.object == object })
	}
	output := make([]SourcePin, 0, len(specs))
	for _, spec := range specs {
		index := slices.IndexFunc(declarations, func(d declaration) bool {
			return d.kind == "function" && d.symbol == spec.function
		})
		if index < 0 {
			return nil, errors.New("source_root_missing")
		}
		root := declarations[index]
		code, e := formatDeclaration(fs, root.node)
		if e != nil {
			return nil, e
		}
		selected := make([]bool, len(declarations))
		pending := []int{index}
		for len(pending) > 0 {
			current := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if selected[current] {
				continue
			}
			selected[current] = true
			d := declarations[current]
			if name, ok := d.object.(*types.TypeName); ok {
				if named, ok := types.Unalias(name.Type()).(*types.Named); ok {
					for k := 0; k < named.NumMethods(); k++ {
						n := lookup(named.Method(k))
						if n < 0 {
							return nil, errors.New("source_method_closure_missing")
						}
						pending = append(pending, n)
					}
				}
			}
			missing := false
			ast.Inspect(d.node, func(node ast.Node) bool {
				ident, ok := node.(*ast.Ident)
				if !ok {
					return true
				}
				object := info.Uses[ident]
				if object == nil || object.Pkg() != pkg {
					return true
				}
				if variable, ok := object.(*types.Var); ok && variable.IsField() {
					return true
				}
				n := lookup(object)
				if n >= 0 {
					pending = append(pending, n)
					return true
				}
				if object.Parent() == pkg.Scope() {
					missing = true
				}
				if function, ok := object.(*types.Func); ok {
					signature, _ := function.Type().(*types.Signature)
					if signature != nil && signature.Recv() != nil {
						missing = true
					}
				}
				return true
			})
			if missing {
				return nil, errors.New("source_authored_closure_missing")
			}
		}
		pin := SourcePin{ID: spec.id, Prototype: spec.prototype, Core: spec.core, Function: spec.function, CodeSHA256: sum(code), NormalizedCodeSHA256: normalizedDeclarationSHA(code)}
		pin.StandardImports = slices.Clone(imports)
		slices.Sort(pin.StandardImports)
		for i, d := range declarations {
			if !selected[i] {
				continue
			}
			raw, e := formatDeclaration(fs, d.node)
			if e != nil {
				return nil, e
			}
			component := ComponentPin{ID: "typedbehavior/" + d.file + ":" + d.kind + ":" + d.symbol, Kind: d.kind, SHA256: sum(raw)}
			// Renamed helpers/methods connect. Pure type shapes or scalar enum
			// layouts alone do not establish a copied behavioral template.
			if d.kind == "function" || d.kind == "method" {
				component.NormalizedBehaviorSHA256 = normalizedDeclarationSHA(raw)
			}
			pin.Components = append(pin.Components, component)
		}
		slices.SortFunc(pin.Components, func(a, b ComponentPin) int { return strings.Compare(a.ID, b.ID) })
		var bundle bytes.Buffer
		for _, c := range pin.Components {
			bundle.WriteString(c.ID)
			bundle.WriteByte(0)
			bundle.WriteString(c.SHA256)
			bundle.WriteByte(0)
		}
		for _, p := range pin.StandardImports {
			bundle.WriteString("std:" + p)
			bundle.WriteByte(0)
		}
		bundle.WriteString("go1.27.1;standard-library-infrastructure-v1")
		pin.BundleSHA256 = sum(bundle.Bytes())
		output = append(output, pin)
	}
	return output, nil
}

// A trimpath binary has no compiled GOROOT. Resolve export metadata explicitly
// from an installed, version-matched Go tool instead of mutating build.Default.
// This belongs only to offline provenance work, never runtime hint ranking.
func sourceImporter(fs *token.FileSet, imports []string) (types.Importer, error) {
	if len(imports) == 0 {
		return importer.ForCompiler(fs, "gc", nil), nil
	}
	goBin, e := exec.LookPath("go")
	if e != nil {
		return nil, errors.New("source_go_tool_unavailable")
	}
	queryDir, e := os.MkdirTemp("", "riido-source-metadata-")
	if e != nil {
		return nil, errors.New("source_go_metadata_unavailable")
	}
	defer os.RemoveAll(queryDir)
	var env []string
	for _, value := range os.Environ() {
		name, _, _ := strings.Cut(value, "=")
		if !slices.Contains([]string{"GOROOT", "GOOS", "GOARCH", "GOTOOLCHAIN", "GOPROXY", "GOSUMDB", "GOWORK", "GOENV", "GOFLAGS", "CGO_ENABLED", "GOMAXPROCS", "GOMEMLIMIT", "GOCACHEPROG", "GOEXPERIMENT"}, name) {
			env = append(env, value)
		}
	}
	env = append(env, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=-p=1", "CGO_ENABLED=0", "GOMAXPROCS=1", "GOMEMLIMIT=256MiB")
	command := func(bin string, args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bin, args...)
		// Bound inherited pipes after the Go parent exits or is canceled.
		// This is not a host-wide process-tree resource guarantee.
		cmd.WaitDelay = time.Second
		cmd.Dir, cmd.Env, cmd.Stderr = queryDir, env, io.Discard
		var output boundedMetadata
		cmd.Stdout = &output
		if cmd.Run() != nil || output.exceeded {
			return nil, errors.New("source_go_metadata_unavailable")
		}
		return output.Bytes(), nil
	}
	metadata, e := command(goBin, "env", "-json", "GOROOT", "GOMODCACHE", "GOVERSION")
	if e != nil {
		return nil, e
	}
	var tool struct{ GOROOT, GOMODCACHE, GOVERSION string }
	if json.Unmarshal(metadata, &tool) != nil {
		return nil, errors.New("source_go_metadata_unavailable")
	}
	if tool.GOVERSION == "go1.27.1" {
		goBin = filepath.Join(tool.GOROOT, "bin", "go")
	} else {
		// Go's cached official toolchain can coexist with an older PATH Go.
		// Locate only the exact local version; never request a download.
		goBin = filepath.Join(tool.GOMODCACHE, "golang.org", "toolchain@v0.0.1-go1.27.1."+runtime.GOOS+"-"+runtime.GOARCH, "bin", "go")
	}
	version, e := command(goBin, "env", "GOVERSION")
	if e != nil || strings.TrimSpace(string(version)) != "go1.27.1" {
		return nil, errors.New("source_go_tool_unpinned")
	}
	args := append([]string{"list", "-deps", "-export", "-json"}, imports...)
	raw, e := command(goBin, args...)
	if e != nil {
		return nil, e
	}
	var exports []struct{ path, file string }
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var pkg struct {
			ImportPath, Export string
			Standard           bool
			Error              json.RawMessage
		}
		e := decoder.Decode(&pkg)
		if e == io.EOF {
			break
		}
		if e != nil || !pkg.Standard || len(pkg.Error) > 0 || len(exports) >= 512 {
			return nil, errors.New("source_go_export_metadata_invalid")
		}
		if pkg.Export != "" {
			exports = append(exports, struct{ path, file string }{pkg.ImportPath, pkg.Export})
		}
	}
	lookup := func(path string) (io.ReadCloser, error) {
		n := slices.IndexFunc(exports, func(x struct{ path, file string }) bool { return x.path == path })
		if n < 0 {
			return nil, errors.New("source_go_export_missing")
		}
		f, e := os.Open(exports[n].file)
		if e != nil {
			return nil, errors.New("source_go_export_missing")
		}
		info, e := f.Stat()
		if e != nil || !info.Mode().IsRegular() || info.Size() > 32<<20 {
			f.Close()
			return nil, errors.New("source_go_export_invalid")
		}
		return f, nil
	}
	return importer.ForCompiler(fs, "gc", lookup), nil
}

type boundedMetadata struct {
	buffer   bytes.Buffer
	exceeded bool
}

func (b *boundedMetadata) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedMetadata) Write(raw []byte) (int, error) {
	if b.buffer.Len()+len(raw) > 16<<20 {
		b.exceeded = true
		return 0, errors.New("source_go_metadata_bounds")
	}
	return b.buffer.Write(raw)
}

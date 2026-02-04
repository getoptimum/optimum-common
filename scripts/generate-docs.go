package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	pkgDir  = "pkg"
	docsDir = "docs/pkg"
)

func main() {
	if err := generateDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Documentation generated successfully")
}

func generateDocs() error {
	// Walk pkg/ directory
	return filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_fuzz_test.go") {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Parse and generate docs for this file
		return processFile(path)
	})
}

func processFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	// Skip if no package declaration
	if node.Name == nil {
		return nil
	}

	// Extract package documentation
	pkgDoc := extractPackageDoc(node)

	// Extract exported items
	funcs := extractFunctions(node, fset)
	types := extractTypes(node, fset)
	consts := extractConstants(node, fset)
	vars := extractVariables(node, fset)

	// Generate markdown
	markdown := generateMarkdown(filePath, node.Name.Name, pkgDoc, funcs, types, consts, vars)

	// Determine output path
	relPath, err := filepath.Rel(pkgDir, filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	outputPath := filepath.Join(docsDir, strings.TrimSuffix(relPath, ".go")+".md")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(markdown), 0644)
}

func extractPackageDoc(node *ast.File) string {
	if node.Doc != nil {
		return strings.TrimSpace(node.Doc.Text())
	}
	return ""
}

func extractFunctions(node *ast.File, fset *token.FileSet) []funcInfo {
	var funcs []funcInfo

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name == nil || !d.Name.IsExported() {
				continue
			}
			if d.Recv != nil {
				continue // Skip methods, they'll be included with types
			}

			funcs = append(funcs, funcInfo{
				Name:    d.Name.Name,
				Doc:     extractComment(d.Doc),
				Signature: formatSignature(d, fset),
			})
		}
	}

	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].Name < funcs[j].Name
	})

	return funcs
}

func extractTypes(node *ast.File, fset *token.FileSet) []typeInfo {
	var types []typeInfo

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}

				// Extract methods for this type
				methods := extractMethods(node, ts.Name.Name, fset)

				// Format type definition
				typeDef := formatType(ts.Type, fset)
				// Store type params separately for rendering
				typeParams := ""
				if ts.TypeParams != nil {
					typeParams = formatTypeParams(ts.TypeParams, fset)
				}
				// Combine: Name[params] TypeDef
				typeStr := ts.Name.Name + typeParams + " " + typeDef

				types = append(types, typeInfo{
					Name:    ts.Name.Name,
					Doc:     extractComment(d.Doc),
					Type:    typeStr,
					Methods: methods,
				})
			}
		}
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	return types
}

func extractMethods(node *ast.File, typeName string, fset *token.FileSet) []funcInfo {
	var methods []funcInfo

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil || len(d.Recv.List) == 0 {
				continue
			}
			if d.Name == nil || !d.Name.IsExported() {
				continue
			}

			// Check if this method belongs to the type
			recvType := formatType(d.Recv.List[0].Type, fset)
			if strings.Contains(recvType, typeName) || strings.Contains(recvType, "*"+typeName) {
				methods = append(methods, funcInfo{
					Name:      d.Name.Name,
					Doc:       extractComment(d.Doc),
					Signature: formatSignature(d, fset),
				})
			}
		}
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return methods
}

func extractConstants(node *ast.File, fset *token.FileSet) []constInfo {
	var consts []constInfo

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.CONST {
				continue
			}
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if !name.IsExported() {
						continue
					}
					consts = append(consts, constInfo{
						Name:  name.Name,
						Doc:   extractComment(d.Doc),
						Value: formatValue(vs, fset),
					})
				}
			}
		}
	}

	sort.Slice(consts, func(i, j int) bool {
		return consts[i].Name < consts[j].Name
	})

	return consts
}

func extractVariables(node *ast.File, fset *token.FileSet) []varInfo {
	var vars []varInfo

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.VAR {
				continue
			}
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if !name.IsExported() {
						continue
					}
					vars = append(vars, varInfo{
						Name:  name.Name,
						Doc:   extractComment(d.Doc),
						Value: formatValue(vs, fset),
					})
				}
			}
		}
	}

	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Name < vars[j].Name
	})

	return vars
}

func extractComment(comment *ast.CommentGroup) string {
	if comment == nil {
		return ""
	}
	return strings.TrimSpace(comment.Text())
}

func formatSignature(fn *ast.FuncDecl, fset *token.FileSet) string {
	var buf strings.Builder
	buf.WriteString("func ")
	if fn.Recv != nil {
		buf.WriteString("(")
		for i, recv := range fn.Recv.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			for j, name := range recv.Names {
				if j > 0 {
					buf.WriteString(", ")
				}
				if name != nil {
					buf.WriteString(name.Name)
					buf.WriteString(" ")
				}
			}
			buf.WriteString(formatType(recv.Type, fset))
		}
		buf.WriteString(") ")
	}
	buf.WriteString(fn.Name.Name)
	
	// Add type parameters if present
	if fn.Type.TypeParams != nil {
		buf.WriteString(formatTypeParams(fn.Type.TypeParams, fset))
	}
	
	buf.WriteString(formatFuncType(fn.Type, fset))
	return buf.String()
}

func formatTypeParams(tp *ast.FieldList, fset *token.FileSet) string {
	if tp == nil || len(tp.List) == 0 {
		return ""
	}
	var buf strings.Builder
	buf.WriteString("[")
	for i, field := range tp.List {
		if i > 0 {
			buf.WriteString(", ")
		}
		for j, name := range field.Names {
			if j > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(name.Name)
		}
		if field.Type != nil {
			buf.WriteString(" ")
			// Handle constraint with ~ operator
			if tilde, ok := field.Type.(*ast.BinaryExpr); ok && tilde.Op.String() == "~" {
				buf.WriteString("~")
				buf.WriteString(formatType(tilde.Y, fset))
			} else {
				buf.WriteString(formatType(field.Type, fset))
			}
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func formatFuncType(ft *ast.FuncType, fset *token.FileSet) string {
	var buf strings.Builder
	buf.WriteString("(")
	if ft.Params != nil {
		for i, param := range ft.Params.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			for j, name := range param.Names {
				if j > 0 {
					buf.WriteString(", ")
				}
				if name != nil {
					buf.WriteString(name.Name)
					buf.WriteString(" ")
				}
			}
			buf.WriteString(formatType(param.Type, fset))
		}
	}
	buf.WriteString(")")
	if ft.Results != nil {
		buf.WriteString(" ")
		if len(ft.Results.List) > 1 {
			buf.WriteString("(")
		}
		for i, result := range ft.Results.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			for j, name := range result.Names {
				if j > 0 {
					buf.WriteString(", ")
				}
				if name != nil {
					buf.WriteString(name.Name)
					buf.WriteString(" ")
				}
			}
			buf.WriteString(formatType(result.Type, fset))
		}
		if len(ft.Results.List) > 1 {
			buf.WriteString(")")
		}
	}
	return buf.String()
}

func formatType(expr ast.Expr, fset *token.FileSet) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return formatType(t.X, fset) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X, fset)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatType(t.Elt, fset)
		}
		return "[" + formatType(t.Len, fset) + "]" + formatType(t.Elt, fset)
	case *ast.MapType:
		return "map[" + formatType(t.Key, fset) + "]" + formatType(t.Value, fset)
	case *ast.ChanType:
		dir := ""
		switch t.Dir {
		case ast.SEND:
			dir = "chan<- "
		case ast.RECV:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + formatType(t.Value, fset)
	case *ast.FuncType:
		return formatFuncType(t, fset)
	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return "interface{}"
		}
		return "interface{...}"
	case *ast.StructType:
		if t.Fields == nil || len(t.Fields.List) == 0 {
			return "struct{}"
		}
		return "struct{...}"
	case *ast.Ellipsis:
		return "..." + formatType(t.Elt, fset)
	case *ast.BasicLit:
		return t.Value
	case *ast.IndexExpr:
		return formatType(t.X, fset) + "[" + formatType(t.Index, fset) + "]"
	case *ast.IndexListExpr:
		var indices []string
		for _, index := range t.Indices {
			indices = append(indices, formatType(index, fset))
		}
		return formatType(t.X, fset) + "[" + strings.Join(indices, ", ") + "]"
	case *ast.BinaryExpr:
		// Handle expressions like "1 * time.Minute"
		return formatType(t.X, fset) + " " + t.Op.String() + " " + formatType(t.Y, fset)
	default:
		// Fallback: try to get source
		return "..."
	}
}

func formatValue(vs *ast.ValueSpec, fset *token.FileSet) string {
	if vs.Values == nil || len(vs.Values) == 0 {
		return ""
	}
	if len(vs.Values) == 1 {
		return formatType(vs.Values[0], fset)
	}
	var values []string
	for _, v := range vs.Values {
		values = append(values, formatType(v, fset))
	}
	return strings.Join(values, ", ")
}

func generateMarkdown(filePath, pkgName, pkgDoc string, funcs []funcInfo, types []typeInfo, consts []constInfo, vars []varInfo) string {
	var buf strings.Builder

	// File header
	fileName := filepath.Base(filePath)
	buf.WriteString(fmt.Sprintf("# Package: %s\n\n", pkgName))
	buf.WriteString(fmt.Sprintf("**File:** `%s`\n\n", fileName))

	// Package documentation
	if pkgDoc != "" {
		buf.WriteString(pkgDoc)
		buf.WriteString("\n\n")
	}

	// Functions section
	if len(funcs) > 0 {
		buf.WriteString("## Functions\n\n")
		for i, fn := range funcs {
			buf.WriteString(fmt.Sprintf("### %s\n\n", fn.Name))
			buf.WriteString("```go\n")
			buf.WriteString(fn.Signature)
			buf.WriteString("\n```\n\n")
			if fn.Doc != "" {
				buf.WriteString(fn.Doc)
				buf.WriteString("\n\n")
			}
			if i < len(funcs)-1 || len(types) > 0 || len(consts) > 0 || len(vars) > 0 {
				buf.WriteString("---\n\n")
			}
		}
	}

	// Types section
	if len(types) > 0 {
		buf.WriteString("## Types\n\n")
		for i, typ := range types {
			buf.WriteString(fmt.Sprintf("### %s\n\n", typ.Name))
			buf.WriteString("```go\n")
			buf.WriteString(fmt.Sprintf("type %s", typ.Type))
			buf.WriteString("\n```\n\n")
			if typ.Doc != "" {
				buf.WriteString(typ.Doc)
				buf.WriteString("\n\n")
			}

			// Methods
			if len(typ.Methods) > 0 {
				buf.WriteString("**Methods:**\n\n")
				for _, method := range typ.Methods {
					buf.WriteString(fmt.Sprintf("- `%s` - ", method.Name))
					if method.Doc != "" {
						// Take first line of doc as summary
						lines := strings.Split(method.Doc, "\n")
						buf.WriteString(lines[0])
					}
					buf.WriteString("\n")
				}
				buf.WriteString("\n")
			}
			if i < len(types)-1 || len(consts) > 0 || len(vars) > 0 {
				buf.WriteString("---\n\n")
			}
		}
	}

	// Constants section
	if len(consts) > 0 {
		buf.WriteString("## Constants\n\n")
		for i, c := range consts {
			buf.WriteString(fmt.Sprintf("### %s\n\n", c.Name))
			buf.WriteString("```go\n")
			buf.WriteString(fmt.Sprintf("const %s", c.Name))
			if c.Value != "" {
				buf.WriteString(" = " + c.Value)
			}
			buf.WriteString("\n```\n\n")
			if c.Doc != "" {
				buf.WriteString(c.Doc)
				buf.WriteString("\n\n")
			}
			if i < len(consts)-1 || len(vars) > 0 {
				buf.WriteString("---\n\n")
			}
		}
	}

	// Variables section
	if len(vars) > 0 {
		buf.WriteString("## Variables\n\n")
		for i, v := range vars {
			buf.WriteString(fmt.Sprintf("### %s\n\n", v.Name))
			buf.WriteString("```go\n")
			buf.WriteString(fmt.Sprintf("var %s", v.Name))
			if v.Value != "" {
				buf.WriteString(" = " + v.Value)
			}
			buf.WriteString("\n```\n\n")
			if v.Doc != "" {
				buf.WriteString(v.Doc)
				buf.WriteString("\n\n")
			}
			if i < len(vars)-1 {
				buf.WriteString("---\n\n")
			}
		}
	}

	result := buf.String()
	// Trim trailing newlines (but ensure file ends with single newline)
	result = strings.TrimRight(result, "\n")
	if result != "" {
		result += "\n"
	}
	return result
}

type funcInfo struct {
	Name      string
	Doc       string
	Signature string
}

type typeInfo struct {
	Name    string
	Doc     string
	Type    string
	Methods []funcInfo
}

type constInfo struct {
	Name  string
	Doc   string
	Value string
}

type varInfo struct {
	Name  string
	Doc   string
	Value string
}

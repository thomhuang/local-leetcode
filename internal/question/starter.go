package question

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const notImplemented = `panic("not implemented")`

// MakeStarterCompilable turns a raw LeetCode starter snippet into Go that
// compiles in this repository. LeetCode hides two things from the snippet: it
// supplies definitions for node types that only show up in a doc comment, and
// its judge tolerates functions whose body is empty even when they return a
// value. Locally those are "undefined" and "missing return" errors, so we add
// the commented definitions back and give empty value-returning bodies a panic.
func MakeStarterCompilable(source string) string {
	if strings.TrimSpace(source) == "" {
		return source
	}
	source = insertCommentedTypeDefinitions(source)
	source = insertNotImplementedPanics(source)
	return source
}

// insertNotImplementedPanics fills any empty function body that must return a
// value with a panic so the file type-checks. Bodies that return nothing are
// left alone; they are already valid.
func insertNotImplementedPanics(source string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return source
	}

	var offsets []int
	ast.Inspect(file, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Body == nil || len(fn.Body.List) != 0 {
			return true
		}
		if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
			return true
		}
		offsets = append(offsets, fset.Position(fn.Body.Lbrace).Offset)
		return true
	})

	// Insert from the end so earlier offsets stay valid.
	for i := len(offsets) - 1; i >= 0; i-- {
		at := offsets[i] + 1
		source = source[:at] + " " + notImplemented + " " + source[at:]
	}
	return source
}

type commentedType struct {
	name string
	code string
}

// insertCommentedTypeDefinitions copies struct definitions that LeetCode leaves
// commented out (TreeNode, ListNode, Node, ...) into real code after the
// package clause. Definitions that are already declared are skipped.
func insertCommentedTypeDefinitions(source string) string {
	blocks := extractCommentedTypes(source)
	if len(blocks) == 0 {
		return source
	}

	declared := declaredTypes(source)
	var additions strings.Builder
	for _, block := range blocks {
		if declared[block.name] {
			continue
		}
		additions.WriteString(block.code)
		additions.WriteString("\n")
	}
	if additions.Len() == 0 {
		return source
	}
	return insertAfterFirstLine(source, additions.String())
}

// declaredTypes lists the type names a file actually declares. Comments are
// ignored so a commented-out definition does not count as a declaration.
func declaredTypes(source string) map[string]bool {
	names := make(map[string]bool)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return names
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				names[typeSpec.Name.Name] = true
			}
		}
	}
	return names
}

func extractCommentedTypes(source string) []commentedType {
	lines := strings.Split(source, "\n")
	var blocks []commentedType

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "* type ") {
			continue
		}

		name := typeNameFromLine(trimmed)
		if name == "" {
			continue
		}

		var code strings.Builder
		depth := 0
		for j := i; j < len(lines); j++ {
			content := stripCommentPrefix(lines[j])
			code.WriteString(content)
			code.WriteString("\n")
			depth += strings.Count(content, "{") - strings.Count(content, "}")
			if depth <= 0 {
				break
			}
		}
		blocks = append(blocks, commentedType{name: name, code: code.String()})
	}
	return blocks
}

func typeNameFromLine(line string) string {
	rest := strings.TrimPrefix(line, "* type ")
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// stripCommentPrefix removes the leading "* " that LeetCode adds to every line
// of a definition comment.
func stripCommentPrefix(line string) string {
	idx := strings.Index(line, "*")
	if idx < 0 {
		return strings.TrimSpace(line)
	}
	return strings.TrimPrefix(line[idx+1:], " ")
}

func insertAfterFirstLine(source, addition string) string {
	idx := strings.Index(source, "\n")
	if idx < 0 {
		return source + "\n\n" + addition
	}
	return source[:idx+1] + "\n" + addition + source[idx+1:]
}

package question

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestMakeStarterCompilableAddsPanicToValueReturningBody(t *testing.T) {
	source := "package contains_duplicate\n\nfunc containsDuplicate(nums []int) bool {\n    \n}\n"

	got := MakeStarterCompilable(source)
	if !strings.Contains(got, notImplemented) {
		t.Fatalf("MakeStarterCompilable() = %q, want a panic body", got)
	}
	assertParses(t, got)
}

func TestMakeStarterCompilableLeavesVoidBody(t *testing.T) {
	source := "package min_stack\n\ntype MinStack struct {\n    \n}\n\nfunc (this *MinStack) Push(value int)  {\n    \n}\n"

	got := MakeStarterCompilable(source)
	if strings.Contains(got, notImplemented) {
		t.Fatalf("MakeStarterCompilable() = %q, want void body untouched", got)
	}
}

func TestMakeStarterCompilableUncommentsTreeNode(t *testing.T) {
	source := "package balanced_binary_tree\n\n" +
		"/**\n" +
		" * Definition for a binary tree node.\n" +
		" * type TreeNode struct {\n" +
		" *     Val int\n" +
		" *     Left *TreeNode\n" +
		" *     Right *TreeNode\n" +
		" * }\n" +
		" */\n" +
		"func isBalanced(root *TreeNode) bool {\n    \n}\n"

	got := MakeStarterCompilable(source)
	if !strings.Contains(got, "\ntype TreeNode struct {") {
		t.Fatalf("MakeStarterCompilable() = %q, want a real TreeNode definition", got)
	}
	if !strings.Contains(got, notImplemented) {
		t.Fatalf("MakeStarterCompilable() = %q, want a panic body", got)
	}
	// The definition must land after the package clause, not before it.
	if strings.Index(got, "package ") > strings.Index(got, "\ntype TreeNode struct {") {
		t.Fatalf("MakeStarterCompilable() put the definition before the package clause: %q", got)
	}
	assertParses(t, got)
}

func TestMakeStarterCompilableDoesNotDuplicateDeclaredType(t *testing.T) {
	source := "package x\n\ntype TreeNode struct {\n    Val int\n}\n\n" +
		"/**\n" +
		" * type TreeNode struct {\n" +
		" *     Val int\n" +
		" * }\n" +
		" */\n" +
		"func f(root *TreeNode) bool {\n}\n"

	got := MakeStarterCompilable(source)
	if strings.Count(got, "\ntype TreeNode struct {") != 1 {
		t.Fatalf("MakeStarterCompilable() = %q, want exactly one TreeNode definition", got)
	}
}

func assertParses(t *testing.T, source string) {
	t.Helper()
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "starter.go", source, parser.ParseComments); err != nil {
		t.Fatalf("generated starter does not parse: %v\n%s", err, source)
	}
}

// Ensure the helper does not leave an empty value-returning body behind.
func TestMakeStarterCompilableNoEmptyResultBodies(t *testing.T) {
	source := "package find_median_from_data_stream\n\ntype MedianFinder struct {\n}\n\n" +
		"func Constructor() MedianFinder {\n    \n}\n\n" +
		"func (this *MedianFinder) AddNum(num int)  {\n    \n}\n\n" +
		"func (this *MedianFinder) FindMedian() float64 {\n    \n}\n"

	got := MakeStarterCompilable(source)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "starter.go", got, parser.ParseComments)
	if err != nil {
		t.Fatalf("generated starter does not parse: %v", err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 && len(fn.Body.List) == 0 {
			t.Fatalf("function %s still has an empty result body", fn.Name.Name)
		}
		return true
	})
}

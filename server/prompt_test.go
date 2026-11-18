package main

import "testing"

func TestPrepareSubmission(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		packageName string
		want        string
	}{
		{
			name:        "CRLF source",
			source:      "package subsets\r\n\r\nfunc subsets(nums []int) [][]int {}\r\n",
			packageName: "subsets",
			want:        "func subsets(nums []int) [][]int {}\r\n",
		},
		{
			name:        "LF source",
			source:      "package word_search\n\nfunc exist(board [][]byte, word string) bool {}\n",
			packageName: "word_search",
			want:        "func exist(board [][]byte, word string) bool {}\n",
		},
		{
			name:        "snippet without package",
			source:      "func subsets(nums []int) [][]int {}\n",
			packageName: "subsets",
			want:        "func subsets(nums []int) [][]int {}\n",
		},
		{
			name: "provided TreeNode",
			source: "package balanced_binary_tree\r\n\r\ntype TreeNode struct {\r\n\tVal   int\r\n\tLeft  *TreeNode\r\n\tRight *TreeNode\r\n}\r\n\r\n" +
				"/**\r\n * type TreeNode struct {\r\n *     Val int\r\n * }\r\n */\r\ntype helper struct{}\r\n\r\nfunc isBalanced(root *TreeNode) bool {}\r\n",
			packageName: "balanced_binary_tree",
			want:        "/**\r\n * type TreeNode struct {\r\n *     Val int\r\n * }\r\n */\r\ntype helper struct{}\r\n\r\nfunc isBalanced(root *TreeNode) bool {}\r\n",
		},
		{
			name:        "different package",
			source:      "package other\n\nfunc subsets(nums []int) [][]int {}\n",
			packageName: "subsets",
			want:        "package other\n\nfunc subsets(nums []int) [][]int {}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := prepareSubmission(test.source, test.packageName)
			if got != test.want {
				t.Fatalf("prepareSubmission() = %q, want %q", got, test.want)
			}
		})
	}
}
func TestParseCookies(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantSession string
		wantCsrf    string
		wantOk      bool
	}{
		{
			name:        "both cookies with extras",
			raw:         "gr_user_id=abc; csrftoken=tok123; LEETCODE_SESSION=sess456; _ga=xyz",
			wantSession: "sess456",
			wantCsrf:    "tok123",
			wantOk:      true,
		},
		{
			name:        "value that contains equals sign",
			raw:         "LEETCODE_SESSION=a.b==; csrftoken=t",
			wantSession: "a.b==",
			wantCsrf:    "t",
			wantOk:      true,
		},
		{
			name:        "leading Cookie header prefix and quotes",
			raw:         `Cookie: "LEETCODE_SESSION=sess789; csrftoken=csrf999; cf_clearance=clear1"`,
			wantSession: "sess789",
			wantCsrf:    "csrf999",
			wantOk:      true,
		},
		{name: "missing session", raw: "csrftoken=t", wantOk: false},
		{name: "garbage", raw: "not a cookie", wantOk: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session, csrf, ok := parseCookies(test.raw)
			if ok != test.wantOk {
				t.Fatalf("ok = %v, want %v", ok, test.wantOk)
			}
			if ok && (session != test.wantSession || csrf != test.wantCsrf) {
				t.Fatalf("got (%q, %q), want (%q, %q)", session, csrf, test.wantSession, test.wantCsrf)
			}
		})
	}
}

func TestNormalizeCookieHeaderPreservesExtraCookies(t *testing.T) {
	header, session, csrf, ok := normalizeCookieHeader("cookie: cf_clearance=abc; LEETCODE_SESSION=s1; csrftoken=c1\r\n")
	if !ok || session != "s1" || csrf != "c1" {
		t.Fatalf("normalizeCookieHeader() = (%q, %q, %q, %v)", header, session, csrf, ok)
	}
	want := "cf_clearance=abc; LEETCODE_SESSION=s1; csrftoken=c1"
	if header != want {
		t.Fatalf("header = %q, want %q", header, want)
	}
}

func TestParseSingleCookieValue(t *testing.T) {
	if got := parseSingleCookieValue("raw_token_value", "LEETCODE_SESSION"); got != "raw_token_value" {
		t.Fatalf("bare value = %q, want raw_token_value", got)
	}
	if got := parseSingleCookieValue("csrftoken=tok123;", "csrftoken"); got != "tok123" {
		t.Fatalf("named value = %q, want tok123", got)
	}
	if got := parseSingleCookieValue("other=tok123", "csrftoken"); got != "" {
		t.Fatalf("wrong name = %q, want empty", got)
	}
	if got := parseSingleCookieValue("not a cookie", "LEETCODE_SESSION"); got != "" {
		t.Fatalf("spaces = %q, want empty", got)
	}
}

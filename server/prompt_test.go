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

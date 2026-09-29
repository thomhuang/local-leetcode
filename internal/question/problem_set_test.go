package question

import "testing"

func TestPackageName(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want string
	}{
		{name: "hyphens become underscores", slug: "two-sum", want: "two_sum"},
		{name: "leading digits move to the end", slug: "01-matrix", want: "matrix01"},
		{name: "single leading digit", slug: "3sum", want: "sum3"},
		{name: "trailing suffix", slug: "permutations-ii", want: "permutations_ii"},
		{name: "short slug", slug: "powx-n", want: "powx_n"},
		{name: "already valid", slug: "subsets", want: "subsets"},
		{name: "keyword gets a suffix", slug: "type", want: "type_"},
		{name: "unknown character becomes underscore", slug: "a.b", want: "a_b"},
		{name: "empty slug falls back", slug: "", want: "problem"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := PackageName(test.slug); got != test.want {
				t.Fatalf("PackageName(%q) = %q, want %q", test.slug, got, test.want)
			}
		})
	}
}

func TestParseProblemSet(t *testing.T) {
	valid := `{
		"name": "NeetCode 150",
		"slug": "neetcode-150",
		"categories": [
			{"name": "Arrays & Hashing", "problems": [
				{"id": 217, "title": "Contains Duplicate", "slug": "contains-duplicate", "difficulty": "Easy"}
			]},
			{"name": "Two Pointers", "problems": [
				{"id": 15, "title": "3Sum", "slug": "3sum", "difficulty": "Medium"}
			]}
		]
	}`

	set, err := ParseProblemSet([]byte(valid))
	if err != nil {
		t.Fatalf("ParseProblemSet() error = %v", err)
	}
	if set.Name != "NeetCode 150" || set.Slug != "neetcode-150" {
		t.Fatalf("unexpected set header: %+v", set)
	}
	if problems := set.Problems(); len(problems) != 2 || problems[1].Slug != "3sum" {
		t.Fatalf("Problems() = %+v, want 2 with second slug 3sum", problems)
	}
}

func TestParseProblemSetRejectsInvalid(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "malformed json", raw: `{"slug": "x"`},
		{name: "missing slug", raw: `{"name": "x", "categories": [{"name": "c", "problems": [{"slug": "a"}]}]}`},
		{name: "no categories", raw: `{"name": "x", "slug": "x", "categories": []}`},
		{name: "empty category", raw: `{"slug": "x", "categories": [{"name": "c", "problems": []}]}`},
		{name: "problem without slug", raw: `{"slug": "x", "categories": [{"name": "c", "problems": [{"title": "t"}]}]}`},
		{
			name: "duplicate slug",
			raw:  `{"slug": "x", "categories": [{"name": "c", "problems": [{"slug": "a"}, {"slug": "a"}]}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseProblemSet([]byte(test.raw)); err == nil {
				t.Fatalf("ParseProblemSet(%s) error = nil, want an error", test.raw)
			}
		})
	}
}

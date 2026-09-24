package solution

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCheckSolutionResponseSubmissionFields(t *testing.T) {
	data := []byte(`{
		"status_code": 11,
		"run_success": true,
		"last_testcase": "[2,7,11,15]\n9",
		"expected_output": "[0,1]",
		"runtime_error": "index out of range"
	}`)

	var got CheckSolutionResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.LastTestcase != "[2,7,11,15]\n9" {
		t.Errorf("LastTestcase = %q", got.LastTestcase)
	}
	if got.ExpectedOutput != "[0,1]" {
		t.Errorf("ExpectedOutput = %q", got.ExpectedOutput)
	}
	if got.RuntimeError != "index out of range" {
		t.Errorf("RuntimeError = %q", got.RuntimeError)
	}
}

func TestCheckSolutionResponseAnswerCanBeStringOrArray(t *testing.T) {
	tests := []struct {
		name string
		json string
		want StringList
	}{
		{name: "array", json: `["[0,1]", "[1,0]"]`, want: StringList{"[0,1]", "[1,0]"}},
		{name: "single string", json: `"[0,1]"`, want: StringList{"[0,1]"}},
		{name: "null", json: `null`, want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response struct {
				CodeAnswer StringList `json:"code_answer"`
			}
			if err := json.Unmarshal([]byte(`{"code_answer":`+test.json+`}`), &response); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(response.CodeAnswer, test.want) {
				t.Errorf("CodeAnswer = %#v, want %#v", response.CodeAnswer, test.want)
			}
		})
	}
}

func TestIsFinal(t *testing.T) {
	tests := map[string]bool{
		Pending:   false,
		Started:   false,
		"SUCCESS": true,
		"FAILURE": true,
		"":        true,
	}
	for state, want := range tests {
		if got := IsFinal(state); got != want {
			t.Errorf("IsFinal(%q) = %v, want %v", state, got, want)
		}
	}
}

package solution

import "testing"

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

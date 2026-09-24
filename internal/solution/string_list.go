package solution

import (
	"encoding/json"
	"fmt"
)

type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*s = nil
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = StringList{single}
		return nil
	}

	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}

	return fmt.Errorf("solution response field must be a string, null, or an array of strings")
}

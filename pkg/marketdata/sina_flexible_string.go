package marketdata

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// SinaJSONString decodes string-like fields from Sina MoneyFlow JSON APIs.
// Some responses use false instead of a string when the display name is absent.
type SinaJSONString string

func (s *SinaJSONString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*s = ""
		return nil
	}
	if data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*s = SinaJSONString(v)
		return nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		*s = SinaJSONString(x)
	case bool:
		*s = ""
	case float64:
		*s = SinaJSONString(fmt.Sprintf("%g", x))
	default:
		*s = SinaJSONString(fmt.Sprint(x))
	}
	return nil
}

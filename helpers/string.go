package helpers

import (
	"fmt"
)

func CutString(s string, minLen uint) string {
	if minLen == 0 {
		return s
	}

	r := []rune(s)
	return string(r[:min(uint(len(r)), minLen)])
}

func DefineString(arg any) string {
	switch v := arg.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return ""
	case *int:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	case *int32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	case *int64:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

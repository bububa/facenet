package imageutil

// IsHex returns true if the string only contains hex numbers, dashes and letters without whitespace.
func IsHex(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < 48 || r > 57) && (r < 97 || r > 102) && (r < 65 || r > 70) && r != 45 {
			return false
		}
	}

	return true
}

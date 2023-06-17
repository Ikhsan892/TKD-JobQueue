package utils

func GetAutoWidth(value interface{}) int {
	oneChar := 3
	gap := 2
	max := 176
	width := 25
	if _, ok := value.(string); ok {
		width = (len([]rune(value.(string))) * oneChar) + gap
	}

	if width > max {
		return max
	}

	return width
}

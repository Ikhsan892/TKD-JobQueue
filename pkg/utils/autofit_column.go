package utils

func GetAutoWidth(value string) int {
	oneChar := 3
	gap := 2
	max := 176
	width := (len([]rune(value)) * oneChar) + gap

	if width > max {
		return max
	}

	return width
}

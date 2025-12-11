package browser

const hintChars = "fjdkslahgurieowpq"

func generateHints(count int) []string {
	chars := []rune(hintChars)
	hints := make([]string, 0, count)

	if count <= len(chars) {
		for i := range count {
			hints = append(hints, string(chars[i]))
		}
		return hints
	}

	for _, first := range chars {
		for _, second := range chars {
			hints = append(hints, string([]rune{first, second}))
			if len(hints) == count {
				return hints
			}
		}
	}

	return hints
}

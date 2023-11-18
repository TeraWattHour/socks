package helpers

func SwapInnerText(text []rune, start int, end int, replacement []rune) (updated string, offset int) {
	updated = string(text[:start]) + string(replacement) + string(text[end:])
	return updated, len(replacement) - (end - start)
}

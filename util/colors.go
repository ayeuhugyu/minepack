package util

import "strconv"

type Color [3]uint8

// NewColor creates a Color from three uint8 values, ensuring each is between 0 and 255.
func NewColor(r, g, b uint8) Color {
	return Color{r, g, b}
}

func ColorAny(text string, color Color) string {
	return "\033[38;2;" + strconv.Itoa(int(color[0])) + ";" + strconv.Itoa(int(color[1])) + ";" + strconv.Itoa(int(color[2])) + "m" + text + "\033[0m"
}

func ResetColor(text ...string) string {
	var t string
	if len(text) > 0 {
		t = text[0]
	}
	return t + "\033[0m"
}

func formatWithSymbol(text, symbol string, color Color) string {
	return ColorAny(symbol+text, color)
}
func FormatError(text string) string {
	return formatWithSymbol(text, "✘  ", NewColor(220, 120, 120))
}
func FormatSuccess(text string) string {
	return formatWithSymbol(text, "✔  ", NewColor(100, 200, 100))
}
func FormatInfo(text string) string {
	return formatWithSymbol(text, "➜  ", NewColor(80, 120, 220))
}

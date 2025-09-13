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
	return ColorAny(symbol+text, color) + "\033[0m"
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

// MoveCursorUp moves the cursor up by the specified number of lines
func MoveCursorUp(lines int) string {
	return "\033[" + strconv.Itoa(lines) + "A"
}

// ClearLine clears the current line
func ClearLine() string {
	return "\033[2K"
}

// MoveCursorToStartOfLine moves cursor to the beginning of the current line
func MoveCursorToStartOfLine() string {
	return "\033[0G"
}

// OverwritePreviousLine moves cursor up one line, clears it, and positions at start
func OverwritePreviousLine() string {
	return MoveCursorUp(1) + ClearLine() + MoveCursorToStartOfLine()
}

// ClearPromptLines clears multiple lines above cursor to clean up prompt artifacts
func ClearPromptLines(lines int) string {
	result := ""
	for i := 0; i < lines; i++ {
		result += MoveCursorUp(1) + ClearLine()
	}
	result += MoveCursorToStartOfLine()
	return result
}

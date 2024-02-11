package helpers

import (
	"fmt"
	"strings"
)

func FixedWidth(s string, size int) string {
	if len(s) > size {
		return s[:size]
	}
	return s + strings.Repeat(" ", size-len(s))
}

type Color string

const (
	RED    Color = "red"
	GREEN        = "green"
	YELLOW       = "yellow"
	BLUE         = "blue"
)

func Colorize(content string, color Color) string {
	colorNumber := 0
	switch color {
	case RED:
		colorNumber = 31
	case GREEN:
		colorNumber = 32
	case YELLOW:
		colorNumber = 33
	case BLUE:
		colorNumber = 34
	}
	return fmt.Sprintf("\033[1;%dm%s\033[00m", colorNumber, content)
}

func Bold(content string) string {
	return fmt.Sprintf("\033[1m%s\033[00m", content)
}

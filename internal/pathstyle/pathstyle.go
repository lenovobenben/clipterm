package pathstyle

import (
	"errors"
	"strings"
	"unicode"
)

const (
	Native  = "native"
	Windows = "windows"
	WSL     = "wsl"
)

var ErrUnsupported = errors.New("unsupported path style")

func Transform(path, style string) (string, error) {
	switch normalizedStyle(style) {
	case "", Native, Windows:
		return path, nil
	case WSL:
		return windowsPathToWSL(path), nil
	default:
		return "", ErrUnsupported
	}
}

func normalizedStyle(style string) string {
	return strings.ToLower(strings.TrimSpace(style))
}

func windowsPathToWSL(path string) string {
	if len(path) < 3 {
		return path
	}

	drive := rune(path[0])
	if !isASCIILetter(drive) || path[1] != ':' || !isWindowsSeparator(path[2]) {
		return path
	}

	rest := strings.ReplaceAll(path[3:], "\\", "/")
	return "/mnt/" + strings.ToLower(string(drive)) + "/" + rest
}

func isASCIILetter(r rune) bool {
	return unicode.IsLetter(r) && r <= unicode.MaxASCII
}

func isWindowsSeparator(b byte) bool {
	return b == '\\' || b == '/'
}

package upgrade

import "strings"

const (
	ansiReset      = "\x1b[0m"
	ansiBoldRed    = "\x1b[1;31m"
	ansiBoldGreen  = "\x1b[1;32m"
	ansiCyan       = "\x1b[36m"
	ansiRed        = "\x1b[31m"
	ansiGreen      = "\x1b[32m"
	ansiDimDefault = "\x1b[2m"
)

func Colorize(diff string) string {
	if diff == "" {
		return ""
	}

	endsWithNewline := strings.HasSuffix(diff, "\n")
	body := diff
	if endsWithNewline {
		body = body[:len(body)-1]
	}

	var b strings.Builder
	b.Grow(len(diff) + 32)

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if code := ansiForLine(line); code != "" {
			b.WriteString(code)
			b.WriteString(line)
			b.WriteString(ansiReset)
		} else {
			b.WriteString(line)
		}
		if i < len(lines)-1 || endsWithNewline {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func ansiForLine(line string) string {
	switch {
	case strings.HasPrefix(line, "--- "):
		return ansiBoldRed
	case strings.HasPrefix(line, "+++ "):
		return ansiBoldGreen
	case strings.HasPrefix(line, "@@ "):
		return ansiCyan
	case strings.HasPrefix(line, "\\ "):
		return ansiDimDefault
	}
	if idx := strings.Index(line, diffSeparator); idx >= 0 {
		after := idx + len(diffSeparator)
		if after < len(line) {
			switch line[after] {
			case '-':
				return ansiRed
			case '+':
				return ansiGreen
			}
		}
	}
	return ""
}

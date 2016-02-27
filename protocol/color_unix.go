// +build darwin freebsd netbsd openbsd linux

package protocol

const (
	colorFrontBlack   = "\x1b[30m"
	colorFrontRed     = "\x1b[31m"
	colorFrontGreen   = "\x1b[32m"
	colorFrontYellow  = "\x1b[33m"
	colorFrontBlue    = "\x1b[34m"
	colorFrontMagenta = "\x1b[35m"
	colorFrontCyan    = "\x1b[36m"
	colorFrontWhite   = "\x1b[37m"
	colorBgBlack      = "\x1b[40m"
	colorBgRed        = "\x1b[41;30m"
	colorBgGreen      = "\x1b[42;30m"
	colorBgYellow     = "\x1b[43;30m"
	colorBgBlue       = "\x1b[44;30m"
	colorBgMagenta    = "\x1b[45;30m"
	colorBgCyan       = "\x1b[46;30m"
	colorBgWhite      = "\x1b[47;30m"
	colorReset        = "\x1b[0m"
)

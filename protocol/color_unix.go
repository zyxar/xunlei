// +build darwin freebsd netbsd openbsd linux

package protocol

const (
	color_front_black   = "\x1b[30m"
	color_front_red     = "\x1b[31m"
	color_front_green   = "\x1b[32m"
	color_front_yellow  = "\x1b[33m"
	color_front_blue    = "\x1b[34m"
	color_front_magenta = "\x1b[35m"
	color_front_cyan    = "\x1b[36m"
	color_front_white   = "\x1b[37m"
	color_bg_black      = "\x1b[40m"
	color_bg_red        = "\x1b[41;30m"
	color_bg_green      = "\x1b[42;30m"
	color_bg_yellow     = "\x1b[43;30m"
	color_bg_blue       = "\x1b[44;30m"
	color_bg_magenta    = "\x1b[45;30m"
	color_bg_cyan       = "\x1b[46;30m"
	color_bg_white      = "\x1b[47;30m"
	color_reset         = "\x1b[0m"
)

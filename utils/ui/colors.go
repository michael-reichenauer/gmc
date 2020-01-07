package ui

type Color int

// Foreground text colors
const (
	CBlack Color = iota + 30
	CRed
	CGreen
	CYellow
	CBlue
	CMagenta
	CCyan
	CWhite
	CGray
	CDark
	CRedDk
	CGreenDk
	CYellowDk
	CBlueDk
	CMagentaDk
	CCyanDk
)

var colors = map[Color]string{
	CBlack:   "\033[30;3m",
	CRed:     "\033[31;1m",
	CBlue:    "\033[34;1m",
	CMagenta: "\033[35;1m",
	CCyan:    "\033[36;1m",
	CWhite:   "\033[37;2m",
	CGray:    "\033[37;3m",
	CDark:    "\033[30;1m",
	CGreen:   "\033[32;1m",
	CYellow:  "\033[33;1m",

	CRedDk:     "\033[31;3m",
	CBlueDk:    "\033[34;3m",
	CMagentaDk: "\033[35;3m",
	CCyanDk:    "\033[36;3m",
	CGreenDk:   "\033[32;3m",
	CYellowDk:  "\033[33;3m",
}

func ColorRune(color Color, r rune) string {
	return colors[color] + string(r) + colorEnd
}
func ColorText(color Color, text string) string {
	return colors[color] + text + colorEnd
}

const (
	colorEnd       = "\033[0m"
	colorWhite     = "\033[37;2m"
	colorGray      = "\033[37;3m"
	colorDark      = "\033[30;1m"
	colorRed       = "\033[31;1m"
	colorRedDk     = "\033[31;3m"
	colorGreen     = "\033[32;1m"
	colorGreenDk   = "\033[32;3m"
	colorYellow    = "\033[33;1m"
	colorYellowDk  = "\033[33;3m"
	colorBlue      = "\033[34;1m"
	colorBlueDk    = "\033[34;3m"
	colorMagenta   = "\033[35;1m"
	colorMagentaDk = "\033[35;3m"
	colorCyan      = "\033[36;1m"
	colorCyanDk    = "\033[36;3m"
)

func Magenta(text string) string {
	return colorMagenta + text + colorEnd
}
func MagentaDk(text string) string {
	return colorMagentaDk + text + colorEnd
}

func Gray(text string) string {
	return colorGray + text + colorEnd
}

func Dark(text string) string {
	return colorDark + text + colorEnd
}

func White(text string) string {
	return colorWhite + text + colorEnd
}

func Red(text string) string {
	return colorRed + text + colorEnd
}
func RedDk(text string) string {
	return colorRedDk + text + colorEnd
}

func Green(text string) string {
	return colorGreen + text + colorEnd
}
func GreenDk(text string) string {
	return colorGreenDk + text + colorEnd
}
func Yellow(text string) string {
	return colorYellow + text + colorEnd
}
func YellowDk(text string) string {
	return colorYellowDk + text + colorEnd
}
func Blue(text string) string {
	return colorBlue + text + colorEnd
}
func BlueDk(text string) string {
	return colorBlueDk + text + colorEnd
}
func Cyan(text string) string {
	return colorCyan + text + colorEnd
}
func CyanDk(text string) string {
	return colorCyanDk + text + colorEnd
}

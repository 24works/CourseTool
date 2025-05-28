package asnicolor

// ANSI 顏色代碼常量
const (
	// 重置所有屬性
	Reset = "\033[0m"

	// 標準前景顏色
	Black  = "\033[30m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m" // 新增紫色
	Cyan   = "\033[36m"
	White  = "\033[37m"

	// 亮色前景顏色
	BrightBlack  = "\033[90m"
	BrightRed    = "\033[91m"
	BrightGreen  = "\033[92m"
	BrightYellow = "\033[93m"
	BrightBlue   = "\033[94m"
	BrightPurple = "\033[95m"
	BrightCyan   = "\033[96m"
	BrightWhite  = "\033[97m"

	// 標準背景顏色
	BgBlack  = "\033[40m"
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"
	BgWhite  = "\033[47m"

	// 亮色背景顏色
	BgBrightBlack  = "\033[100m"
	BgBrightRed    = "\033[101m"
	BgBrightGreen  = "\033[102m"
	BgBrightYellow = "\033[103m"
	BgBrightBlue   = "\033[104m"
	BgBrightPurple = "\033[105m"
	BgBrightCyan   = "\033[106m"
	BgBrightWhite  = "\033[107m"

	// 文本樣式
	Bold      = "\033[1m" // 粗體
	Dim       = "\033[2m" // 弱化/變暗
	Italic    = "\033[3m" // 斜體 (不廣泛支持)
	Underline = "\033[4m" // 下劃線
	Blink     = "\033[5m" // 閃爍 (不廣泛支持)
	Reverse   = "\033[7m" // 反轉前景和背景顏色
	Hidden    = "\033[8m" // 隱藏文本
	Strike    = "\033[9m" // 刪除線 (不廣泛支持)

	// 更多文本樣式
	Fraktur             = "\033[20m" // 哥特體 (不廣泛支持)
	DoublyUnderline     = "\033[21m" // 雙下劃線 (有時作為粗體重置)
	NormalIntensity     = "\033[22m" // 正常強度 (重置粗體和弱化)
	NoUnderline         = "\033[24m" // 無下劃線 (重置下劃線)
	NoBlink             = "\033[25m" // 無閃爍 (重置閃爍)
	NoReverse           = "\033[27m" // 無反轉 (重置反轉)
	Reveal              = "\033[28m" // 顯示文本 (重置隱藏)
	NoStrike            = "\033[29m" // 無刪除線 (重置刪除線)
	Framed              = "\033[51m" // 帶框
	Encircled           = "\033[52m" // 帶圓圈
	Overline            = "\033[53m" // 上劃線
	NoFramedOrEncircled = "\033[54m" // 重置帶框或帶圓圈
	NoOverline          = "\033[55m" // 重置上劃線

	// 256 色和真彩色 (通常需要更複雜的序列，這裡只列出基本類型)
	// 這些通常通過 "\033[38;5;{color_code}m" (前景) 或 "\033[48;5;{color_code}m" (背景) 使用
	// 或 "\033[38;2;{r};{g};{b}m" (前景) 或 "\033[48;2;{r};{g};{b}m" (背景)
)

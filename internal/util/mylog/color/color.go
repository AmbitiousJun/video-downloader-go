package color

// 日志颜色输出常量
const (
	Blue   = "\x1b[38;2;90;156;248m"
	Green  = "\x1b[38;2;126;192;80m"
	Yellow = "\x1b[38;2;220;165;80m"
	Red    = "\x1b[38;2;228;116;112m"
	Purple = "\x1b[38;2;160;186;250m"

	reset = "\x1b[0m"
)

// ToBlue 将字符串转成蓝色
func ToBlue(str string) string {
	return Blue + str + reset
}

// ToGreen 将字符串转成绿色
func ToGreen(str string) string {
	return Green + str + reset
}

// ToYellow 将字符串转成黄色
func ToYellow(str string) string {
	return Yellow + str + reset
}

// ToRed 将字符串转成红色
func ToRed(str string) string {
	return Red + str + reset
}

// ToPurple 将字符串转成紫色
func ToPurple(str string) string {
	return Purple + str + reset
}

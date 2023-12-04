package strings

// GetStrLen 获取字符串长度
func GetStrLen(str string) int {
	return len([]rune(str))
}

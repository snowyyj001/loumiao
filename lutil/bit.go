package lutil

func Bit(n int) int {
	return 1 << n
}

func BitAnd(n1 int, n2 int) int {
	return n1 & n2
}

func BitOr(n1 int, n2 int) int {
	return n1 | n2
}

func HasBit(n1 int, n2 int) bool {
	return (n1 & n2) != 0
}

func HasNotBit(n1 int, n2 int) bool {
	return (n1 & n2) == 0
}

func EqualBit(val int, flag int) bool {
	return (val & flag) == flag
}

func Bit64(n int64) int64 {
	return 1 << n
}

func Bit64NAnd(n1 int64, n2 int64) int64 {
	return n1 & (^n2)
}

func Bit64And(n1 int64, n2 int64) int64 {
	return n1 & n2
}

func Bit64Or(n1 int64, n2 int64) int64 {
	return n1 | n2
}

func HasBit64(n1 int64, n2 int64) bool {
	return (n1 & n2) != 0
}

func EqualBit64(val int64, flag int64) bool {
	return (val & flag) == flag
}

// AddFlag 给n1添加n2 bit位标识
func AddFlag(n1 int64, n2 int64) int64 {
	return Bit64Or(n1, Bit64(n2))
}

// RemoveFlag 给n1去掉n2 bit位标识
func RemoveFlag(n1 int64, n2 int64) int64 {
	return Bit64NAnd(n1, Bit64(n2))
}

// HasFlag n1是否有n2 bit位标识
func HasFlag(n1 int64, n2 int64) bool {
	return Bit64And(n1, Bit64(n2)) != 0
}

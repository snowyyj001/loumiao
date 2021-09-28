package util

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

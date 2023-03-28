package lutil

/*
Diffie-Hellman密钥交换算法
百度百科
https://baike.baidu.com/item/Diffie-Hellman/9827194?fr=aladdin
*/
/*
const(
	DH_Q = 97
	DH_A = 5
	dh_XA = 36
	dh_XB = 58
)
*/
const (
	DH_Q  = 193
	DH_A  = 5
	dh_XA = 123
	dh_XB = 132
)

var (
	DH_YA = 0
	DH_YB = 0
	DH_KA = 0
	DH_KB = 0
)

func Calc_DH_Ya() { //a^XA mod q
	leftA := 1
	leftB := 1
	for i := 1; i < dh_XA; i++ {
		if leftA < DH_Q {
			leftA *= DH_A
			i++
		}
		if leftA >= DH_Q {
			leftB *= leftA % DH_Q
			leftA = 1
		}
		if leftB >= DH_Q {
			leftB %= DH_Q
		}
	}
	DH_YA = (leftA * leftB) % DH_Q
}

func Calc_DH_Yb() { //a^XB mod q
	leftA := 1
	leftB := 1
	for i := 1; i < dh_XB; i++ {
		if leftA < DH_Q {
			leftA *= DH_A
			i++
		}
		if leftA >= DH_Q {
			leftB *= leftA % DH_Q
			leftA = 1
		}
		if leftB >= DH_Q {
			leftB %= DH_Q
		}
	}
	DH_YB = (leftA * leftB) % DH_Q
}

func Calc_DH_A_k() { //YB^XA mod q
	leftA := 1
	leftB := 1
	for i := 1; i < dh_XA; i++ {
		if leftA < DH_Q {
			leftA *= DH_YB
			i++
		}
		if leftA >= DH_Q {
			leftB *= leftA % DH_Q
			leftA = 1
		}
		if leftB >= DH_Q {
			leftB %= DH_Q
		}
	}
	DH_KA = (leftA * leftB) % DH_Q
}

func Calc_DH_B_k() { //YA^XB mod q
	leftA := 1
	leftB := 1
	for i := 1; i < dh_XB; i++ {
		if leftA < DH_Q {
			leftA *= DH_YA
			i++
		}
		if leftA >= DH_Q {
			leftB *= leftA % DH_Q
			leftA = 1
		}
		if leftB >= DH_Q {
			leftB %= DH_Q
		}
	}
	DH_KB = (leftA * leftB) % DH_Q
}

// 最终dh的K值
func GetDHK() int {
	return DH_KA
}

func init() {
	Calc_DH_Yb()
	Calc_DH_A_k()
}

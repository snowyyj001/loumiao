package util

//随机数生成算法-伪随机-32位
type MTRand struct {
	mt    [624]uint32 //624 * 32 - 31 = 19937
	index int
}

func (self *MTRand) setSeed(seed int) {
	self.mt[0] = uint32(seed)
	//对数组的其它元素进行初始化
	for i := 1; i < 624; i++ {
		t := 1812433253*(self.mt[i-1]^(self.mt[i-1]>>30)) + uint32(i)
		self.mt[i] = t & 0xffffffff //取最后的32位赋给MT[i]
	}
}

func (self *MTRand) generate() {
	for i := 0; i < 624; i++ {
		// 2^31 = 0x80000000
		// 2^31-1 = 0x7fffffff
		y := (self.mt[i] & 0x80000000) + (self.mt[(i+1)%624] & 0x7fffffff)
		self.mt[i] = self.mt[(i+397)%624] ^ (y >> 1)
		if y&1 != 0 {
			self.mt[i] ^= 2567483615
		}
	}
}

//生成一个随机数
func (self *MTRand) Random() int {
	if self.index == 0 {
		self.generate()
	}
	y := self.mt[self.index]
	y = y ^ (y >> 11)                //y右移11个bit
	y = y ^ ((y << 7) & 2636928640)  //y左移7个bit与2636928640相与，再与y进行异或
	y = y ^ ((y << 15) & 4022730752) //y左移15个bit与4022730752相与，再与y进行异或
	y = y ^ (y >> 18)                //y右移18个bit再与y进行异或
	self.index = (self.index + 1) % 624

	return int(y)
}

//生成一个随机数[0, n)
func (self *MTRand) Rand(n int) int {
	return self.Random() % n
}

//构建一个随机数生成器
func NewMTRand(seed int) *MTRand {
	mt := new(MTRand)
	mt.setSeed(seed)
	return mt
}

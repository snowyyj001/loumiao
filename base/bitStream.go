package base

import "encoding/binary"

//----------------bitsream---------------
//for example
//buf := make([]byte, 256)
//var bitstream base.BitStream
//bitstream.BuildPacketStream(buf, 256)
//bitstream.WriteInt(1000, 16)
// or
//bitstream := NewBitStream(buf)
//----------------------------------------

const (
	Bit8       = 8
	Bit16      = 16
	Bit32      = 32
	Bit64      = 64
	MAX_PACKET = 5 * 1024 * 1024 //5MB
)

type (
	BitStream struct {
		dataPtr        []byte
		bitNum         int
		flagNum        int
		tailFlag       bool
		bufSize        int
		bitsLimite     int
		error          bool
		maxReadBitNum  int
		maxWriteBitNum int
	}

	IBitStream interface {
		BuildPacketStream([]byte, int) bool
		setBuffer([]byte, int, int)
		GetBuffer() []byte
		GetBytePtr() []byte
		GetReadByteSize() int
		GetCurPos() int
		GetPosition() int
		GetStreamSize() int
		SetPosition(int) bool
		Reset()
		clear()
		resize() bool

		WriteBits([]byte, int)
		ReadBits(int) []byte
		WriteInt(int, int)
		ReadInt(int) int
		ReadFlag() bool
		WriteFlag(bool) bool
		WriteString(string)
		ReadString() string

		WriteInt64(int64, int)
		ReadInt64(int) int64
		WriteFloat(float32)
		ReadFloat() float32
		WriteFloat64(float64)
		ReadFloat64() float64
	}
)

func (self *BitStream) BuildPacketStream(buffer []byte, writeSize int) bool {
	if writeSize <= 0 {
		return false
	}

	self.setBuffer(buffer, writeSize, -1)
	self.SetPosition(0)
	return true
}

func (self *BitStream) setBuffer(bufPtr []byte, size int, maxSize int) {
	self.dataPtr = bufPtr
	self.bitNum = 0
	self.flagNum = 0
	self.tailFlag = false
	self.bufSize = size
	self.maxReadBitNum = size << 3
	if maxSize < 0 {
		maxSize = size
	}
	self.maxWriteBitNum = maxSize << 3
	self.bitsLimite = size
	self.error = false
}

func (self *BitStream) Reset() {
	self.setBuffer(self.dataPtr, self.bufSize, -1)
	self.SetPosition(0)
}

func (self *BitStream) GetBuffer() []byte {
	return self.dataPtr[0:self.GetPosition()]
}

func (self *BitStream) GetBytePtr() []byte {
	return self.dataPtr[self.GetPosition():]
}

func (self *BitStream) GetReadByteSize() int {
	return (self.maxReadBitNum >> 3) - self.GetPosition()
}

func (self *BitStream) GetCurPos() int {
	return self.bitNum
}

func (self *BitStream) GetPosition() int {
	return (self.bitNum + 7) >> 3
}

func (self *BitStream) GetStreamSize() int {
	return self.bufSize
}

func (self *BitStream) SetPosition(pos int) bool {
	Assert(pos == 0 || self.flagNum == 0, "不正确的setPosition调用")
	if pos != 0 && self.flagNum != 0 {
		return false
	}

	self.bitNum = pos << 3
	self.flagNum = 0
	return true
}

func (self *BitStream) clear() {
	var buff []byte
	buff = make([]byte, self.bufSize)
	self.dataPtr = buff
}

func (self *BitStream) resize() bool {
	//fmt.Println("BitStream Resize")
	self.dataPtr = append(self.dataPtr, make([]byte, self.bitsLimite)...)
	size := self.bitsLimite * 2
	if size <= 0 || size >= MAX_PACKET*2 {
		return false
	}
	self.bufSize = size
	self.maxReadBitNum = size << 3
	self.maxWriteBitNum = size << 3
	self.bitsLimite = size
	return true
}

//@bitCount: 8 16 24 32 40 48 64 ...
func (self *BitStream) WriteBits(bitPtr []byte, bitCount int) {
	if bitCount == 0 {
		return
	}

	if self.tailFlag {
		self.error = true
		Assert(false, "Out of range write")
		return
	}

	if (bitCount & 0x7) != 0 {
		bitCount = (bitCount & ^0x7) + 8
	}

	for bitCount+self.bitNum > self.maxWriteBitNum {
		if !self.resize() {
			self.error = true
			Assert(false, "Out of range write")
			return
		}
	}

	bitNum := self.bitNum >> 3
	byteCount := (bitCount + 7) >> 3
	copy(self.dataPtr[bitNum:], bitPtr[:byteCount])
	/*for i, v := range bitPtr[:byteCount] {
		self.dataPtr[bitNum+i] = v
	}*/
	self.bitNum += bitCount
}

func (self *BitStream) ReadBits(bitCount int) []byte {
	if bitCount == 0 {
		return []byte{}
	}

	if self.tailFlag {
		self.error = true
		Assert(false, "Out of range read")
		return []byte{}
	}

	if (bitCount & 0x7) != 0 {
		bitCount = (bitCount & ^0x7) + 8
	}

	for bitCount+self.bitNum > self.maxReadBitNum {
		if !self.resize() {
			self.error = true
			Assert(false, "Out of range read")
			return []byte{}
		}
	}

	byteCount := (bitCount + 7) >> 3
	bitNum := self.bitNum >> 3
	stPtr := self.dataPtr[bitNum : bitNum+byteCount]
	self.bitNum += bitCount
	return stPtr
}

func (self *BitStream) WriteInt8(value int) {
	self.WriteBits(IntToBytes(value), Bit8)
}

func (self *BitStream) ReadInt8() int {
	var ret int
	buf := self.ReadBits(Bit8)
	ret = BytesToInt(buf)
	return ret
}

func (self *BitStream) WriteInt16(value int) {
	self.WriteBits(IntToBytes(value), Bit16)
}

func (self *BitStream) ReadInt16() int {
	var ret int
	buf := self.ReadBits(Bit16)
	ret = BytesToInt(buf)
	return ret
}

func (self *BitStream) WriteInt32(value int) {
	self.WriteBits(IntToBytes(value), Bit32)
}

func (self *BitStream) ReadInt32() int {
	var ret int
	buf := self.ReadBits(Bit32)
	ret = BytesToInt(buf)
	return ret
}

func (self *BitStream) WriteInt(value int, bitCount int) {
	self.WriteBits(IntToBytes(value), bitCount)
}

func (self *BitStream) ReadInt(bitCount int) int {
	var ret int
	buf := self.ReadBits(bitCount)
	ret = BytesToInt(buf)
	if bitCount == Bit32 {
		return int(ret)
	} else {
		ret &= (1 << uint32(bitCount)) - 1
	}

	return int(ret)
}

func (self *BitStream) ReadFlag() bool {
	buf := self.ReadBits(Bit8)
	v := int8(buf[0])
	return v == 1
}

func (self *BitStream) WriteFlag(value bool) bool {
	if value {
		self.WriteBits([]byte{1}, Bit8)
	} else {
		self.WriteBits([]byte{0}, Bit8)
	}
	return value
}

/*
func (self *BitStream) ReadFlag() bool {
	if ((self.flagNum - (self.flagNum>>3)<<3) == 0) && !self.tailFlag {
		self.flagNum = self.bitNum
		if self.bitNum+8 < self.maxReadBitNum {
			self.bitNum += 8
		} else {
			if !self.resize() {
				self.tailFlag = true
			} else {
				self.bitNum += 8
			}
		}
	}

	if self.flagNum+1 > self.maxReadBitNum {
		self.error = true
		Assert(false, "Out of range read")
		return false
	}

	mask := 1 << uint32(self.flagNum&0x7)
	ret := (int(self.dataPtr[(self.flagNum>>3)]) & mask) != 0
	self.flagNum++
	return ret
}

func (self *BitStream) WriteFlag(value bool) bool {
	if ((self.flagNum - (self.flagNum>>3)<<3) == 0) && !self.tailFlag {
		self.flagNum = self.bitNum

		if self.bitNum+8 < self.maxWriteBitNum {
			self.bitNum += 8 //跳开8个用于写flag
		} else {
			if !self.resize() {
				self.tailFlag = true
			} else {
				self.bitNum += 8 //跳开8个用于写flag
			}
		}
	}

	if self.flagNum+1 > self.maxWriteBitNum {
		self.error = true
		Assert(false, "Out of range write")
		return false
	}

	if value {
		self.dataPtr[(self.flagNum >> 3)] |= 1 << uint32(self.flagNum&0x7)
	} else {
		self.dataPtr[(self.flagNum >> 3)] &= ^(1 << uint32(self.flagNum&0x7))
	}

	self.flagNum++
	return value
}*/

func (self *BitStream) ReadString() string {
	if self.ReadFlag() {
		nLen := self.ReadInt(Bit16)
		buf := self.ReadBits(nLen << 3)
		return string(buf)
	}
	return string("")
}

func (self *BitStream) WriteString(value string) {
	buf := []byte(value)
	nLen := len(buf)
	if self.WriteFlag(nLen > 0) {
		self.WriteInt(nLen, Bit16)
		self.WriteBits(buf, nLen<<3)
	}
}

func (self *BitStream) WriteInt64(value int64) {
	self.WriteBits(Int64ToBytes(value, binary.LittleEndian), Bit64)
}

func (self *BitStream) ReadInt64() int64 {
	var ret int64
	buf := self.ReadBits(Bit64)
	ret = BytesToInt64(buf, binary.LittleEndian)
	return ret
}

func (self *BitStream) WriteFloat(value float32) {
	self.WriteBits(Float32ToByte(value), Bit32)
}

func (self *BitStream) ReadFloat() float32 {
	var ret float32
	buf := self.ReadBits(Bit32)
	ret = BytesToFloat32(buf)

	return float32(ret)
}

func (self *BitStream) WriteFloat64(value float64) {
	self.WriteBits(Float64ToByte(value), Bit64)
}

func (self *BitStream) ReadFloat64() float64 {
	var ret float64
	buf := self.ReadBits(Bit64)
	ret = BytesToFloat64(buf)

	return float64(ret)
}

func (self *BitStream) WriteBytes(buf []byte) {
	nLen := len(buf)
	if self.WriteFlag(nLen > 0) {
		self.WriteInt(nLen, Bit16)
		self.WriteBits(buf, nLen<<3)
	}
}

func (self *BitStream) ReadBytes() []byte {
	if self.ReadFlag() {
		nLen := self.ReadInt(Bit16)
		buf := self.ReadBits(nLen << 3)
		return buf
	}
	return []byte{}
}

//根据buff构造一个bitstream，一般用来接收消息
func NewBitStream(buf []byte, nLen int) *BitStream {
	var bitstream BitStream
	if nLen == 0 {
		nLen = len(buf)
	}
	bitstream.BuildPacketStream(buf, nLen)
	return &bitstream
}

//根据buff构造一个bitstream，一般用来接收消息
func NewBitStreamR(buf []byte) *BitStream {
	var bitstream BitStream
	nLen := len(buf)

	bitstream.BuildPacketStream(buf, nLen)
	return &bitstream
}

//构造一个nLen大小的bitstream，一般用来发送消息
func NewBitStreamS(nLen int) *BitStream {
	var bitstream BitStream
	buf := make([]byte, nLen)
	bitstream.BuildPacketStream(buf, nLen)
	return &bitstream
}

//一个字符串占用的字节大小
func BitStrLen(str string) int {
	sz := len(str)
	if sz == 0 {
		return 0
	}
	return sz + 2
}

package gate

type LouMiaoHeartBeat struct {
	Uid   int
	Child int
}

type LouMiaoHandShake struct {
	Uid int
}

type LouMiaoRpcMsg struct {
	ClientId int
	Buffer   []byte
}

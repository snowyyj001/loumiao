package callrpc

import (
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"sync"
)

var (
	This        *CallRpcServer
)

type CallRpcServer struct {
	gorpc.GoRoutineLogic

	mRpcWait sync.Map //map[string]chan interface{}		//igo name -> chan
}

func (self *CallRpcServer) DoInit() bool {
	llog.Infof("%s DoInit", self.Name)
	This = self
	return true
}

func (self *CallRpcServer) DoRegsiter() {
	llog.Infof("%s DoRegsiter", self.Name)

	self.Register("CallRpc", callRpc)
	self.Register("ReqRpcCall", reqRpcCall)
	self.Register("RespRpcCall", respRpcCall)
}

//begin communicate with other nodes
func (self *CallRpcServer) DoStart() {
	llog.Infof("%s DoStart", self.Name)
}
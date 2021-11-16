package callrpc

import (
	"github.com/golang/protobuf/proto"
	"reflect"

	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
	"github.com/snowyyj001/loumiao/msg"
	"github.com/snowyyj001/loumiao/util"
)

func callRpc(igo gorpc.IGoRoutine, data interface{}) interface{} {
	session := data.(string)

	ch, ok := This.mRpcWait.Load(session)
	if ok == false {
		ch = make(chan interface{})
		This.mRpcWait.Store(session, ch)
	}
	resp := <-ch.(chan interface{})
	return resp
}

func respRpcCall(igo gorpc.IGoRoutine, data interface{}) interface{} {
	rpcmsg := data.(*msg.LouMiaoRpcMsg)
	bitstream := base.NewBitStream(rpcmsg.Buffer, len(rpcmsg.Buffer))
	session := bitstream.ReadString()
	respdata := bitstream.GetBytePtr()
	if ch, ok := This.mRpcWait.LoadAndDelete(session); ok {
		ch.(chan interface{}) <- respdata
	} else {
		llog.Errorf("respRpcCall session[%s] is nil", session)
	}
	return nil
}

func reqRpcCall(igo gorpc.IGoRoutine, data interface{}) interface{} {
	req := data.(*gorpc.MM)
	rpcmsg := req.Data.(*msg.LouMiaoRpcMsg)
	bitstream := base.NewBitStream(rpcmsg.Buffer, len(rpcmsg.Buffer))
	session := bitstream.ReadString()
	respdata := bitstream.GetBytePtr()

	resp, ok := igo.CallActor(req.Id, rpcmsg.FuncName, respdata)
	if ok == false {
		return nil
	}
	m := &gorpc.M{Id: int(rpcmsg.SourceId), Name: session}
	m.Param = util.BitOr(define.RPCMSG_FLAG_RESP, define.RPCMSG_FLAG_CALL)
	if reflect.TypeOf(resp).Kind() == reflect.Slice { //bitstream
		orgbuff := resp.([]byte)
		bitstream := base.NewBitStream_1(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBits(orgbuff, base.BytesLen(orgbuff))
		m.Data = bitstream.GetBuffer()
	} else {
		orgbuff, err := message.Pack(resp.(proto.Message))
		if err != nil {
			llog.Errorf("reqRpcCall: func = %s, session = %s", rpcmsg.FuncName, session)
		}
		bitstream := base.NewBitStream_1(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBits(orgbuff, base.BytesLen(orgbuff))
		m.Data = bitstream.GetBuffer()
	}
	gorpc.MGR.Send("GateServer", "SendRpc", m)
	return nil
}

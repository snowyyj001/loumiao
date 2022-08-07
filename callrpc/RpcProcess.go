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
		ch = make(chan []byte)
		This.mRpcWait.Store(session, ch)
	}
	resp := <-ch.(chan []byte)
	return resp
}

func respRpcCall(igo gorpc.IGoRoutine, data interface{}) interface{} {
	rpcmsg := data.(*msg.LouMiaoRpcMsg)
	bitstream := base.NewBitStream(rpcmsg.Buffer, len(rpcmsg.Buffer))
	session := bitstream.ReadString()
	respdata := bitstream.ReadBytes()
	if ch, ok := This.mRpcWait.LoadAndDelete(session); ok {
		ch.(chan []byte) <- respdata
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
	respdata := bitstream.ReadBytes()
	resp, ok := igo.CallActor(req.Id, rpcmsg.FuncName, respdata)
	if ok == false {
		return nil
	}
	m := &gorpc.M{Id: int(rpcmsg.SourceId), Name: session}
	m.Param = util.BitOr(define.RPCMSG_FLAG_RESP, define.RPCMSG_FLAG_CALL)
	if resp == nil { //rpc调用出错了才会返回nil
		bitstream = base.NewBitStreamS(base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(nil)
		m.Data = bitstream.GetBuffer()
	} else if reflect.TypeOf(resp).Kind() == reflect.Slice { //bitstream
		orgbuff := resp.([]byte)
		bitstream = base.NewBitStreamS(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		m.Data = bitstream.GetBuffer()
	} else {
		orgbuff, err := message.Pack(resp.(proto.Message))
		if err != nil {
			llog.Errorf("reqRpcCall: func = %s, session = %s", rpcmsg.FuncName, session)
		}
		bitstream = base.NewBitStreamS(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		m.Data = bitstream.GetBuffer()
	}
	gorpc.MGR.Send("GateServer", "SendRpc", m)
	return nil
}

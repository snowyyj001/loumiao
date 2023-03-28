package callrpc

import (
	"github.com/golang/protobuf/proto"
	"github.com/snowyyj001/loumiao/lbase"
	"github.com/snowyyj001/loumiao/ldefine"
	"github.com/snowyyj001/loumiao/lgate"
	"github.com/snowyyj001/loumiao/lutil"
	"github.com/snowyyj001/loumiao/pbmsg"
	"reflect"

	"github.com/snowyyj001/loumiao/gorpc"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/message"
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
	rpcmsg := data.(*pbmsg.LouMiaoRpcMsg)
	bitstream := lbase.NewBitStream(rpcmsg.Buffer, len(rpcmsg.Buffer))
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
	rpcmsg := req.Data.(*pbmsg.LouMiaoRpcMsg)
	bitstream := lbase.NewBitStream(rpcmsg.Buffer, len(rpcmsg.Buffer))
	session := bitstream.ReadString()
	respdata := bitstream.ReadBytes()
	resp, ok := igo.CallActor(req.Id, rpcmsg.FuncName, respdata)
	if ok == false {
		return nil
	}
	var buffer []byte
	flag := lutil.BitOr(ldefine.RPCMSG_FLAG_RESP, ldefine.RPCMSG_FLAG_CALL)
	if resp == nil { //rpc调用出错了才会返回nil
		bitstream = lbase.NewBitStreamS(lbase.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(nil)
		buffer = bitstream.GetBuffer()
	} else if reflect.TypeOf(resp).Kind() == reflect.Slice { //bitstream
		orgbuff := resp.([]byte)
		bitstream = lbase.NewBitStreamS(len(orgbuff) + lbase.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		buffer = bitstream.GetBuffer()
	} else {
		orgbuff, err := message.Pack(resp.(proto.Message))
		if err != nil {
			llog.Errorf("reqRpcCall: func = %s, session = %s", rpcmsg.FuncName, session)
		}
		bitstream = lbase.NewBitStreamS(len(orgbuff) + lbase.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBytes(orgbuff)
		buffer = bitstream.GetBuffer()
	}
	lgate.SendRpc(int(rpcmsg.SourceId), session, buffer, flag)
	return nil
}

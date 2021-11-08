package callrpc

import (
	"reflect"

	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"
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
	var resp interface{}
	if util.HasBit(int(rpcmsg.Flag), define.RPCMSG_FLAG_PB) { //pb format
		err, _, _, pm := message.Decode(config.SERVER_NODE_UID, respdata, len(respdata))
		if err != nil {
			llog.Errorf("respRpcCall decode msg error: session = %s, func=%s, error=%s ", session, rpcmsg.FuncName, err.Error())
		}
		resp = pm
	} else {
		resp = respdata
	}
	if ch, ok := This.mRpcWait.LoadAndDelete(session); ok {
		ch.(chan interface{}) <- resp
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

	var resp interface{}
	var ok bool
	if util.HasNotBit(int(rpcmsg.Flag), define.RPCMSG_FLAG_PB) { //not pb but bytes
		resp, ok = igo.CallActor(req.Id, rpcmsg.FuncName, respdata)
	} else {
		err, _, _, pm := message.Decode(config.SERVER_NODE_UID, respdata, len(respdata))
		if err != nil {
			llog.Errorf("reqRpcCall decode msg error : target = %s, func=%s, error=%s ", req.Id, rpcmsg.FuncName, err.Error())
			return nil
		}
		resp, ok = igo.CallActor(req.Id, rpcmsg.FuncName, pm)
	}
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
		orgbuff, _ := message.Encode(int(rpcmsg.SourceId), "", resp)
		bitstream := base.NewBitStream_1(len(orgbuff) + base.BitStrLen(session))
		bitstream.WriteString(session)
		bitstream.WriteBits(orgbuff, base.BytesLen(orgbuff))
		m.Data = bitstream.GetBuffer()
		m.Param = util.BitOr(m.Param, define.RPCMSG_FLAG_PB)
	}
	gorpc.MGR.Send("GateServer", "SendRpc", m)
	return nil
}

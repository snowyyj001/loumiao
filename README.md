# loumiao

*just go game framework*

*分布式集群游戏框架，单个节点基于actor模型以实现无锁化编程*

## 服务类型节点如下

ServerType_None = iota //0 物理机控制节点

ServerType_Gate //1 网关，集群

ServerType_Account //2 账号，集群

ServerType_World //3 世界，集群

ServerType_Zone //4 地图，集群

ServerType_DB //5 数据库，集群

ServerType_Log //6 日志，集群

ServerType_Public //7 唯一公共服，单节点

ServerType_WEB_GM //8 web gm 后台，单节点

ServerType_WEB_LOGIN //9 登录，集群

ServerType_RPCGate //10 rpc gate，集群

ServerType_ETCF //11 配置中心，单节点

ServerType_LOGINQUEUE //12 登录排队， 单点

ServerType_Match //13 匹配， 集群

ServerType_Robot //14 机器人

ServerType_WebKeyPoint //15 数据埋点

ServerType_WebCharge //16 web 充值

## 服务内部actor通信

异步send：loumiao.SendActor

例如通知邮件actor更新发送邮件

```
loumiao.SendActor("MailServer", "SendMail", mailItem)
```

同步call：IGoRoutine.CallActor

例如从邮件actor获取邮件信息

```
IGoRoutine.CallActor("MailServer", "loadMailData", req)
```

## 服务间rpc通信

异步send：loumiao.SendRpc

例如通知匹配服的**pvp**排队actor玩家加入匹配队列

```
loumiao.SendRpc("matchserver/pvp.joinMatchRoom", pack, 0)
```

同步call：loumiao.CallRpc

例如从db服务节点加载玩家数据

```
resp, ok := loumiao.CallRpc(this, "dbserver/game.loadUserData", stream.GetBuffer(), 0)
```

## 如何创建一个服务节点

### 1 如何创建一个永久生命周期actor

```
// 创建一个服务，稍后开启
// @name: actor名，唯一
// @sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Prepare(igo gorpc.IGoRoutine, name string, sync bool) {
	igo.SetSync(sync)
	gorpc.MGR.Start(igo, name)
	igo.Register("ServiceHandler", gorpc.ServiceHandler)
}

// 创建一个服务,立即开启
// @name: actor名，唯一
// @sync: 是否异步协程，无状态服务可以是异步协程，有状态服务应该使用同步协程，可以保证协程安全
func Start(igo gorpc.IGoRoutine, name string, sync bool) {
	Prepare(igo, name, sync)
	gorpc.MGR.DoSingleStart(name)
}
```

### 2 创建gate，gate是一个节点的入口，内部服务和外部服务都需要一个gate来和其他服务交互

```
开启对外的gate
watchDog := &lgate.GateServer{ServerType: network.CLIENT_CONNECT}
loumiao.Start(watchDog, "GateServer", false)
开启队内的gate
watchDog = &lgate.GateServer{ServerType: network.SERVER_CONNECT}
loumiao.Start(watchDog, "GateServer", false)

```

### 3 创建game actor，并注册相关消息

注册actor消息

```
GoRoutineLogic.Register
```

注册网络消息

```
loumiao.RegisterNetHandler
```

注册rpc消息

```
loumiao.RegisterRpcHandler
```

将actor注册进actor管理器

```
loumiao.Prepare(new(game.GameServer), "GameServer", false)
```

### 4 开启actor服务

```
loumiao.Run()

```

### 5 短暂生命周期actor

对于玩家player，每个player对应一个actor，这个actor会随着玩家的退出而销毁 使用GoRoutinePool来管理这类会临时创建销毁的actor 例如： 定义玩家actor管理池

```
type AgentMgr struct {
	gorpc.GoRoutinePool
}
```

添加一个玩家

```
func (self *AgentMgr) AddAgent(player *Agent) {
	if player.UserId <= 0 {
		llog.Errorf("AgentMgr.AddAgent error: userid=%d", player.UserId)
		return
	}
	self.DoSingleStart(player, player.UserId, true)
	//上报当前玩家数量
	nodemgr.OnlineNum = Pool.GetAgentNum()
}
```

获取一个玩家

```
func (self *AgentMgr) GetAgent(userid int64) *Agent {
	igo := self.GetRunRoutine(userid) //已经被标记关闭的routine，就认为它已经被删除了
	if igo != nil {
		return igo.(*Agent)
	}
	return nil
}
```

删除一个玩家

```
func (self *AgentMgr) RemoveAgent(userid int64) {
	self.CloseRoutineCleanly(userid) //安全的关闭routine，并不会立即删除

	//上报当前玩家数量
	nodemgr.OnlineNum = Pool.GetAgentNum()
}
```

给玩家actor发送添加道具命令

```
player = AgentMgr.GetAgent(1000)
player.SendActor("addItem", item)
```

## 服务器内部架构

客户端直连gate gate和world zone queue保持socket tcp专线连接，用来转发来自客户端的网络消息，如有其他业务节点也需要接受客户端网络消息，在gate中配置即可

如有帧同步需要，client还需保持与zone的一个udp专线连接，否则依靠gate的网络转发

除login gm keypoint三个web节点外，所有的服务器节点都参与rpc组网

![image](https://github.com/snowyyj001/loumiao/blob/master/doc/%E6%9C%8D%E5%8A%A1%E5%99%A8%E7%BB%93%E6%9E%84.png?raw=true)

## 通用服务器架构

![image](https://github.com/snowyyj001/loumiao/blob/master/doc/%E6%80%BB%E6%9E%84%E5%9B%BE.jpg?raw=true)


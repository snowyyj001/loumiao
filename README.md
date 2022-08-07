### loumiao
just go game framework

master分支：使用的是etcd作为服发现和共享配置，有数据落地功能

etcf分支：使用的是内部etcf模块作为服发现和共享配置，没有数据落地功能，纯内存数据

### 服务类型节点如下

ServerType_None      = iota //0 物理机控制节点

ServerType_Gate             //1 网关，集群

ServerType_Match          //2 匹配，集群

ServerType_World            //3 世界，集群

ServerType_Zone             //4 地图，集群

ServerType_DB               //5 数据库，集群

ServerType_Log              //6 日志，集群

ServerType_Public               //7 唯一公共服，单节点

ServerType_WEB_GM           //8 web gm 后台，单节点

ServerType_WEB_LOGIN        //9 登录，集群

ServerType_RPCGate          //10 rpc gate，集群

ServerType_ETCF             //11 配置中心，单节点

ServerType_LOGINQUEUE        //12 登录排队， 单点
	
ServerType_Robot             //13 机器人


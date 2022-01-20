### loumiao
just go game framework

master分支：使用的是etcd作为服发现和共享配置，有数据落地功能

etcf分支：使用的是内部etcf模块作为服发现和共享配置，没有数据落地功能，纯内存数据

### 服务类型节点如下

ServerType_None      = iota //0 物理机控制节点

ServerType_Gate             //1 网关，集群

ServerType_Account          //2 账号，集群

ServerType_World            //3 世界，集群

ServerType_Zone             //4 地图，集群

ServerType_DB               //5 数据库，集群

ServerType_Log              //6 日志，集群

ServerType_IM               //7 聊天-跨服-公共服，单节点

ServerType_WEB_GM           //8 web gm，单节点

ServerType_WEB_LOGIN        //9 web login，集群

ServerType_RPCGate          //10 rpc gate，单节点

ServerType_ETCF             //11 配置中心，单节点



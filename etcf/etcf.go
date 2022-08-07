package etcf

// NewClientDis 创建服务发现
func NewEtcf() (*EtcfClient, error) {
	return newEtcfClient()
}

//通用发现
//@prefix: 监听key值
//@hanlder: key值变化回调
func (self *EtcfClient) WatchCommon(prefix string, hanlder func(string, string, bool)) error {
	WatchKey(prefix, hanlder)
	return nil
}

func (self *EtcfClient) PutNode() error {
	return PutNode()
}

func (self *EtcfClient) PutStatus() error {
	return PutStatus()
}
func (self *EtcfClient) SetValue(prefix, val string) error {
	return SetValue(prefix, val)
}

func (self *EtcfClient) GetOne(key string) (string, error) {
	return GetOne(key)
}

func (self *EtcfClient) GetAll(prefix string) ([]string, error) {
	return GetAll(prefix)
}

//撤销租约
func (self *EtcfClient) RevokeLease() error {

	return nil
}

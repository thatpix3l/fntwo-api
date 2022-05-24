package cfg

import "strconv"

type Keys struct {
	VmcListenIP          string
	VmcListenPort        int
	WebListenIP          string
	WebListenPort        int
	ModelUpdateFrequency int
	ConfigPath           string
}

func (k Keys) GetVmcSocketAddress() string {
	return k.VmcListenIP + ":" + strconv.Itoa(k.VmcListenPort)
}

func (k Keys) GetWebSocketAddress() string {
	return k.WebListenIP + ":" + strconv.Itoa(k.WebListenPort)
}

package cfg

import "strconv"

type Keys struct {
	VmcListenIP          string // IP address to listen for VMC data
	VmcListenPort        int    // Port to listen for VMC data
	WebListenIP          string // IP address to serve the frontend
	WebListenPort        int    // Port to serve the frontend
	ModelUpdateFrequency int    // Times per second to send the live model data to frontend clients
	ConfigPath           string // Path to config file
}

func (k Keys) GetVmcSocketAddress() string {
	return k.VmcListenIP + ":" + strconv.Itoa(k.VmcListenPort)
}

func (k Keys) GetWebSocketAddress() string {
	return k.WebListenIP + ":" + strconv.Itoa(k.WebListenPort)
}

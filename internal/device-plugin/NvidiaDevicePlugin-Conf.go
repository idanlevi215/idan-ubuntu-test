///
///
/// Define NvidiaDevicePlugin
///
///
///

package deviceplugindana

import (
	"github.com/NVIDIA/go-gpuallocator/gpuallocator"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// Constants to represent the variousdevice list strategies
const (
	DeviceListStrategyEnvvar       = "envvar"
	DeviceListStrategyVolumeMounts = "volume-mounts"
)

// Constants for use by the 'volume-mounts' device list strategy
const (
	deviceListAsVolumeMountsHostPath          = "/dev/null"
	deviceListAsVolumeMountsContainerPathRoot = "/var/run/nvidia-container-devices"
)

// NvidiaDevicePlugin implements the Kubernetes device plugin API
type NvidiaDevicePlugin struct {
	ResourceManager
	resourceName     string
	deviceListEnvvar string
	allocatePolicy   gpuallocator.Policy
	socket           string

	server        *grpc.Server
	cachedDevices []*Device
	health        chan *Device
	stop          chan interface{}
}

// NewNvidiaDevicePlugin returns an initialized NvidiaDevicePlugin
func NewNvidiaDevicePlugin(resourceName string, resourceManager ResourceManager, deviceListEnvvar string, allocatePolicy gpuallocator.Policy, socket string) *NvidiaDevicePlugin {
	return &NvidiaDevicePlugin{
		ResourceManager:  resourceManager,
		resourceName:     resourceName,
		deviceListEnvvar: deviceListEnvvar,
		allocatePolicy:   allocatePolicy,
		socket:           socket,

		// These will be reinitialized every
		// time the plugin server is restarted.
		cachedDevices: nil,
		server:        nil,
		health:        nil,
		stop:          nil,
	}
}

func (m *NvidiaDevicePlugin) initialize() {
	m.cachedDevices = m.Devices()
	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	m.health = make(chan *Device)
	m.stop = make(chan interface{})
}

func (m *NvidiaDevicePlugin) cleanup() {
	close(m.stop)
	m.cachedDevices = nil
	m.server = nil
	m.health = nil
	m.stop = nil
}

// GetDevicePluginOptions returns the values of the optional settings for this plugin
func (m *NvidiaDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	options := &pluginapi.DevicePluginOptions{
		GetPreferredAllocationAvailable: (m.allocatePolicy != nil),
	}
	return options, nil
}

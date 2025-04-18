package core

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/evilsocket/opensnitch/daemon/log"
)

// LoadEbpfModule loads the given eBPF module, from the given path if specified.
// Otherwise t'll try to load the module from several default paths.
func LoadEbpfModule(module, path string) (m *ebpf.Collection, err error) {
	var (
		modulesDir = "/opensnitchd/ebpf"
		paths      = []string{
			fmt.Sprint("/usr/local/lib", modulesDir),
			fmt.Sprint("/usr/lib", modulesDir),
			fmt.Sprint("/etc/opensnitchd"), // Deprecated: will be removed in future versions.
		}
	)

	// if path has been specified, try to load the module from there.
	if path != "" {
		paths = []string{path}
	}

	modulePath := ""
	moduleError := fmt.Errorf(`Module not found (%s) in any of the paths.
You may need to install the corresponding package`, module)

	logLevel := ebpf.LogLevel(0)
	if log.GetLogLevel() == log.DEBUG {
		logLevel = (ebpf.LogLevelBranch | ebpf.LogLevelInstruction | ebpf.LogLevelStats)
	}
	collOpts := ebpf.CollectionOptions{
		Programs: ebpf.ProgramOptions{LogLevel: logLevel},
	}

	for _, p := range paths {
		modulePath = fmt.Sprint(p, "/", module)
		log.Debug("[eBPF] trying to load %s", modulePath)
		if !Exists(modulePath) {
			continue
		}
		specs, err := ebpf.LoadCollectionSpec(modulePath)
		if err != nil {
			log.Error("[eBPF] module specs error: %s", err)
			continue
		}
		m, err := ebpf.NewCollectionWithOptions(specs, collOpts)
		if err != nil {
			log.Error("[eBPF] module collection error: %s", err)
			continue
		}

		log.Info("[eBPF] module loaded: %s", modulePath)
		return m, nil
	}
	moduleError = fmt.Errorf(`
unable to load eBPF module (%s). Your kernel version (%s) might not be compatible.
If this error persists, change process monitor method to 'proc'`, module, GetKernelVersion())

	return m, moduleError
}

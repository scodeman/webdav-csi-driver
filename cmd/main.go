package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scodeman/webdav-csi-driver/pkg/driver"

	"k8s.io/klog"
)

var (
	conf driver.Config
)

func main() {
	var version bool

	// Parse parameters
	flag.StringVar(&conf.Endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&conf.NodeID, "nodeid", "", "node id")
	flag.StringVar(&conf.SecretPath, "secretpath", "/etc/webdav-csi-dirver", "Secret mount path")

	flag.BoolVar(&version, "version", false, "Print driver version information")

	klog.InitFlags(nil)
	flag.Parse()

	// Handle Version
	if version {
		info, err := driver.GetVersionJSON()

		if err != nil {
			klog.Fatalln(err)
		}

		fmt.Println(info)
		os.Exit(0)
	}

	klog.V(1).Infof("Driver version: %s", driver.GetDriverVersion())

	if conf.NodeID == "" {
		klog.Fatalln("Node ID is not given")
	}

	drv := driver.NewDriver(&conf)
	if err := drv.Run(); err != nil {
		klog.Fatalln(err)
	}

	os.Exit(0)
}

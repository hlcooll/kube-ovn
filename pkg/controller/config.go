package controller

import (
	"flag"
	"github.com/alauda/kube-ovn/pkg/util"
	"os"

	clientset "github.com/alauda/kube-ovn/pkg/client/clientset/versioned"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// Configuration is the controller conf
type Configuration struct {
	BindAddress    string
	OvnNbSocket    string
	OvnNbHost      string
	OvnNbPort      int
	KubeConfigFile string
	KubeClient     kubernetes.Interface
	KubeOvnClient  clientset.Interface

	DefaultLogicalSwitch string
	DefaultCIDR          string
	DefaultGateway       string
	DefaultExcludeIps    string

	ClusterRouter     string
	NodeSwitch        string
	NodeSwitchCIDR    string
	NodeSwitchGateway string

	ClusterTcpLoadBalancer string
	ClusterUdpLoadBalancer string

	PodName      string
	PodNamespace string

	WorkerNum int
	PprofPort int
}

// ParseFlags parses cmd args then init kubeclient and conf
// TODO: validate configuration
func ParseFlags() (*Configuration, error) {
	var (
		argOvnNbSocket    = pflag.String("ovn-nb-socket", "", "The ovn-nb socket file. (If not set use ovn-nb-address)")
		argOvnNbHost      = pflag.String("ovn-nb-host", "0.0.0.0", "The ovn-nb host address. (If not set use ovn-nb-socket)")
		argOvnNbPort      = pflag.Int("ovn-nb-port", 6641, "")
		argKubeConfigFile = pflag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information. If not set use the inCluster token.")

		argDefaultLogicalSwitch = pflag.String("default-ls", "ovn-default", "The default logical switch name, default: ovn-default")
		argDefaultCIDR          = pflag.String("default-cidr", "10.16.0.0/16", "Default CIDR for namespace with no logical switch annotation, default: 10.16.0.0/16")
		argDefaultGateway       = pflag.String("default-gateway", "", "Default gateway for default-cidr, default the first ip in default-cidr")
		argDefaultExcludeIps    = pflag.String("default-exclude-ips", "", "Exclude ips in default switch, default equals to gateway address")

		argClusterRouter     = pflag.String("cluster-router", "ovn-cluster", "The router name for cluster router, default: ovn-cluster")
		argNodeSwitch        = pflag.String("node-switch", "join", "The name of node gateway switch which help node to access pod network, default: join")
		argNodeSwitchCIDR    = pflag.String("node-switch-cidr", "100.64.0.0/16", "The cidr for node switch, default: 100.64.0.0/16")
		argNodeSwitchGateway = pflag.String("node-switch-gateway", "", "The gateway for node switch, default the first ip in node-switch-cidr")

		argClusterTcpLoadBalancer = pflag.String("cluster-tcp-loadbalancer", "cluster-tcp-loadbalancer", "The name for cluster tcp loadbalancer")
		argClusterUdpLoadBalancer = pflag.String("cluster-udp-loadbalancer", "cluster-udp-loadbalancer", "The name for cluster udp loadbalancer")

		argWorkerNum = pflag.Int("worker-num", 3, "The parallelism of each worker, default: 3")
		argPprofPort = pflag.Int("pprof-port", 10660, "The port to get profiling data, default 10660")
	)

	flag.Set("alsologtostderr", "true")

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})

	pflag.CommandLine.AddGoFlagSet(klogFlags)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	config := &Configuration{
		OvnNbSocket:            *argOvnNbSocket,
		OvnNbHost:              *argOvnNbHost,
		OvnNbPort:              *argOvnNbPort,
		KubeConfigFile:         *argKubeConfigFile,
		DefaultLogicalSwitch:   *argDefaultLogicalSwitch,
		DefaultCIDR:            *argDefaultCIDR,
		DefaultGateway:         *argDefaultGateway,
		DefaultExcludeIps:      *argDefaultExcludeIps,
		ClusterRouter:          *argClusterRouter,
		NodeSwitch:             *argNodeSwitch,
		NodeSwitchCIDR:         *argNodeSwitchCIDR,
		NodeSwitchGateway:      *argNodeSwitchGateway,
		ClusterTcpLoadBalancer: *argClusterTcpLoadBalancer,
		ClusterUdpLoadBalancer: *argClusterUdpLoadBalancer,
		WorkerNum:              *argWorkerNum,
		PprofPort:              *argPprofPort,
		PodName:                os.Getenv("POD_NAME"),
		PodNamespace:           os.Getenv("KUBE_NAMESPACE"),
	}

	if config.DefaultGateway == "" {
		gw, err := util.FirstSubnetIP(config.DefaultCIDR)
		if err != nil {
			return nil, err
		}
		config.DefaultGateway = gw
	}

	if config.DefaultExcludeIps == "" {
		config.DefaultExcludeIps = config.DefaultGateway
	}

	if config.NodeSwitchGateway == "" {
		gw, err := util.FirstSubnetIP(config.NodeSwitchCIDR)
		if err != nil {
			return nil, err
		}
		config.NodeSwitchGateway = gw
	}

	err := config.initKubeClient()
	if err != nil {
		return nil, err
	}

	klog.Infof("config is  %v", config)

	return config, nil
}

func (config *Configuration) initKubeClient() error {
	var cfg *rest.Config
	var err error
	if config.KubeConfigFile == "" {
		klog.Infof("no --kubeconfig, use in-cluster kubernetes config")
		cfg, err = rest.InClusterConfig()
		if err != nil {
			klog.Errorf("use in cluster config failed %v", err)
			return err
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", config.KubeConfigFile)
		if err != nil {
			klog.Errorf("use --kubeconfig %s failed %v", config.KubeConfigFile, err)
			return err
		}
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init kubernetes client failed %v", err)
		return err
	}

	kubeOvnClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("init kubeovn client failed %v", err)
		return err
	}

	config.KubeClient = kubeClient
	config.KubeOvnClient = kubeOvnClient
	return nil
}

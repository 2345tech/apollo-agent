package boot

import (
	"flag"
	"fmt"
	"github.com/2345tech/apollo-agent/common"
	"github.com/2345tech/apollo-agent/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	_defaultLogfile    = "./logs/agent.log"
	_defaultConfigFile = "./conf/app.yaml"
	_defaultPprof      = false
)

type Args struct {
	agent      *Agent
	LogFile    *string
	ConfigFile *string
	Pprof      *bool

	helper *helper
}

type helper struct {
	version       *bool
	author        *bool
	convertConfig *bool
}

func NewArg() *Args {
	return &Args{
		helper: &helper{},
	}
}

func (a *Args) Init(agent *Agent) {
	a.agent = agent
	if util.Str("APOLLO_AGENT_SERVER_ADDRESS", "") != "" {
		agent.EnvProfile = true
		stdOut := "/dev/stdout"
		a.LogFile = &stdOut
	} else {
		a.LogFile = flag.String("l", _defaultLogfile, "log string: the log file name with absolute path")
	}
	a.ConfigFile = flag.String("c", _defaultConfigFile, "config string: the config file name with absolute path")
	a.Pprof = flag.Bool("p", _defaultPprof, "pprof bool: open pprof for debug, default http port is 18081")

	a.helper.version = flag.Bool("V", false, "print version")
	a.helper.author = flag.Bool("A", false, "print author")
	a.helper.convertConfig = flag.Bool("convertConfig", false, "convert apolloAgentForPHP config to apollo-agent")

	flag.Parse()

	a.usage()
}

func (a *Args) usage() {
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *a.helper.version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if *a.helper.author {
		fmt.Println(AUTHOR)
		os.Exit(0)
	}

	if *a.helper.convertConfig {
		a.convertConfig()
		os.Exit(0)
	}
}

func (a *Args) convertConfig() {
	oldConfigFile := "/opt/app/apolloAgentForPHP/conf/app.yaml"
	newConfigFile := "/opt/app/apollo-agent/conf/app.yaml"
	if len(flag.Args()) > 0 {
		oldConfigFile = flag.Arg(0)
	}
	if len(flag.Args()) > 1 {
		newConfigFile = flag.Arg(1)
	}

	if convertOldConfigFileToNew(oldConfigFile, newConfigFile) {
		fmt.Println("===================CONVERT OLD CONFIG TO NEW SUCCESS===================")
		return
	}
	fmt.Println("FAILED!FAILED!FAILED!FAILED!FAILED!FAILED!")
}

type oldConfig struct {
	Type        int        `yaml:"type"`
	AllInOne    int        `yaml:"allInOne"`
	LogExpire   int        `yaml:"logExpire"`
	ClusterName string     `yaml:"clusterName"`
	Address     string     `yaml:"address"`
	Ip          string     `yaml:"ip"`
	Configs     []multiApp `yaml:"configs"`
}

type multiApp struct {
	Path      string   `yaml:"path"`
	FileName  string   `yaml:"filename"`
	Syntax    string   `yaml:"syntax"`
	AppId     string   `yaml:"appId"`
	Interval  int      `yaml:"interval"`
	Secret    string   `yaml:"secret"`
	Namespace []string `yaml:"namespace"`
}

func convertOldConfigFileToNew(oldFile, newFile string) bool {
	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		fmt.Println("[ERROR] The configOld config file=" + oldFile + " not exist")
		return false
	} else if os.IsPermission(err) {
		fmt.Println("[ERROR] The configOld config file=" + oldFile + " permission denied")
		return false
	}

	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		fmt.Println("[INFO] The new config NotExist will be create")
	} else if os.IsPermission(err) {
		fmt.Println("[ERROR] The new config file=" + newFile + " permission denied")
		return false
	} else {
		fmt.Println("[INFO] The new config Exist will be copy to " + newFile + ".example")
		if err := util.CopyFile(newFile, newFile+".example"); err != nil {
			fmt.Println("[ERROR] Copy new config to " + newFile + ".example Failed. err:" + err.Error())
			return false
		}
	}

	configOld := oldConfig{}
	oldConfigContent, err := ioutil.ReadFile(oldFile)
	if err != nil {
		fmt.Println("[ERROR] ReadFile old config file:" + oldFile + ", error:" + err.Error())
		return false
	}
	err = yaml.Unmarshal(oldConfigContent, &configOld)
	if err != nil {
		fmt.Println("[ERROR] Unmarshal old config file:" + oldFile + ", error:" + err.Error())
		return false
	}
	configNew := Profile{
		Client: &Client{},
		Server: &Server{},
		Apps:   []*App{},
	}
	if configOld.Type == 1 {
		configNew.Client.Type = common.ModePoll
	} else {
		configNew.Client.Type = common.ModeWatch
	}
	if configOld.AllInOne == 1 {
		configNew.Client.AllInOne = true
	} else {
		configNew.Client.AllInOne = false
	}
	configNew.Client.LogExpire = time.Duration(configOld.LogExpire) * 24 * time.Hour
	configNew.Client.Ip = configOld.Ip

	configNew.Server.Address = configOld.Address
	configNew.Server.Cluster = configOld.ClusterName

	if len(configOld.Configs) > 0 {
		for _, appOld := range configOld.Configs {
			if len(appOld.Namespace) < 1 {
				continue
			}
			path := strings.TrimRight(appOld.Path, string(os.PathSeparator)) + string(os.PathSeparator)
			syntax := _defaultAppSyntax
			if appOld.Syntax != "" {
				syntax = strings.ToLower(appOld.Syntax)
			}
			namespaces := make([]string, 0)
			for _, ns := range appOld.Namespace {
				nsSlice := strings.Split(ns, ",")
				if len(nsSlice) == 1 {
					ns = ns + ".properties"
				}
				namespaces = append(namespaces, ns)
			}
			appNew := App{
				AppId:        appOld.AppId,
				Namespaces:   namespaces,
				Secret:       appOld.Secret,
				Syntax:       syntax,
				PollInterval: time.Duration(appOld.Interval) * time.Second,
				InOneFile:    path + strings.TrimLeft(appOld.FileName, string(os.PathSeparator)),
			}
			configNew.Apps = append(configNew.Apps, &appNew)
		}
	}

	newConfigContent, err := yaml.Marshal(configNew)
	if err != nil {
		fmt.Println("[ERROR] Marshal new config file error:" + err.Error())
		return false
	}

	if err := util.WriteFile(newFile, string(newConfigContent), 0644); err != nil {
		fmt.Println("[ERROR] WriteFile new config file error:" + err.Error())
		return false
	}

	return true
}

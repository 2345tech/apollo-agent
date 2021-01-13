package boot

import (
	"flag"
	"fmt"
	"github.com/2345tech/apollo-agent/util"
	"os"
)

const (
	_defaultLogfile    = "./logs/agent.log"
	_defaultConfigFile = "./conf/app.yaml"
	_defaultPprof      = false
)

type Args struct {
	version    *bool
	author     *bool

	agent      *Agent
	LogFile    *string
	ConfigFile *string
	Pprof      *bool
}

func NewArg() *Args {
	return &Args{}
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
	a.version = flag.Bool("V", false, "print version")
	a.author = flag.Bool("author", false, "print author")
	flag.Parse()

	a.usage()
}

func (a *Args) usage() {
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *a.version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if *a.author {
		fmt.Println(AUTHOR)
		os.Exit(0)
	}
}

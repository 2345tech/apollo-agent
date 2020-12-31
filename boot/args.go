package boot

import (
	"flag"
	"github.com/2345tech/apollo-agent/util"
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
	_ = flag.Set("version", VERSION)
	_ = flag.Set("author", AUTHOR)

	flag.Parse()
}

package boot

import (
	"flag"
	"github.com/2345tech/apollo-agent/util"
)

const (
	_defaultLogfile    = "./logs/agent.log"
	_defaultConfigFile = "./conf/app.yaml"
)

type Args struct {
	agent      *Agent
	LogFile    *string
	ConfigFile *string
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
		a.LogFile = flag.String("log", _defaultLogfile, "the log file name with absolute path")
	}
	a.ConfigFile = flag.String("config", _defaultConfigFile, "the config file name with absolute path")
	_ = flag.Set("version", VERSION)
	_ = flag.Set("author", AUTHOR)

	flag.Parse()
}

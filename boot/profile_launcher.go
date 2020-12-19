package boot

import (
	"fmt"
	"github.com/2345tech/apollo-agent/util"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	_defaultClientType      = "poll"
	_defaultClientAllInOne  = true
	_defaultClientLogExpire = 1 * 24 * time.Hour
	_defaultServerCluster   = "default"
	_defaultAppNamespace    = "application.properties"
	_defaultAppPollInterval = 60 * time.Second
	_defaultAppSyntax       = util.F_ENV
)

type ProfileLauncher struct {
	booted        bool
	agent         *Agent
	watcher       *fsnotify.Watcher
	FromEnvVar    bool
	Profile       *Profile
	ProfileUpdate bool
}

type Profile struct {
	Client *Client `yaml:"client"`
	Server *Server `yaml:"server"`
	Apps   []*App  `yaml:"apps"`
}

type Client struct {
	Type      string        `yaml:"pollOrWatch"`
	AllInOne  bool          `yaml:"allInOne"`
	LogExpire time.Duration `yaml:"logExpire"`
	Ip        string        `yaml:"ip"`
	BeatFreQ  time.Duration `yaml:"beatFreq"`
}

type Server struct {
	Address string `yaml:"address"`
	Cluster string `yaml:"cluster"`
}

type App struct {
	AppId        string        `yaml:"appId"`
	Namespaces   []string      `yaml:"namespace"`
	Secret       string        `yaml:"secret"`
	PollInterval time.Duration `yaml:"pollInterval"`
	InOne        *InOne        `yaml:"inOne"`
}

type InOne struct {
	FileName string `yaml:"filename"`
	Syntax   string `yaml:"syntax"`
}

func NewProfile() *ProfileLauncher {
	return &ProfileLauncher{
		booted:        false,
		FromEnvVar:    false,
		ProfileUpdate: true,
	}
}

func (p *ProfileLauncher) Init(agent *Agent) error {
	p.agent = agent
	agent.ConfigL = p
	if err := p.Parse(); err != nil {
		return fmt.Errorf("[ERROR] " + err.Error())
	}
	agent.LogExpire = p.Profile.Client.LogExpire
	agent.BeatFreQ = p.Profile.Client.BeatFreQ
	return nil
}

func (p *ProfileLauncher) Run() error {
	var err error
	if err = p.Parse(); err != nil {
		return fmt.Errorf("[ERROR] " + err.Error())
	}
	if p.booted {
		return nil
	}
	if p.watcher, err = fsnotify.NewWatcher(); err != nil {
		return fmt.Errorf("[ERROR] " + err.Error())
	}

	go p.watchConfigFile()
	if err = p.watcher.Add(*p.agent.Args.ConfigFile); err != nil {
		return fmt.Errorf("[ERROR] " + err.Error())
	}
	p.booted = true
	return nil
}

func (p *ProfileLauncher) Shutdown() {
	_ = p.watcher.Close()
	p.booted = false
	log.Println("[INFO] ProfileLauncher stopped")
}

func (p *ProfileLauncher) Parse() error {
	if !p.ProfileUpdate {
		return nil
	}
	if p.agent.EnvProfile {
		if err := p.loadEnvVar(); err != nil {
			return err
		}
	} else {
		if err := p.loadConfigFile(); err != nil {
			return err
		}
	}
	p.Profile.wrapper()
	p.ProfileUpdate = false
	return nil
}

func (p *ProfileLauncher) loadEnvVar() error {
	p.Profile.Client.Type = util.Str("APOLLO_AGENT_CLIENT_TYPE", "")
	p.Profile.Client.AllInOne = util.Bool("APOLLO_AGENT_CLIENT_ALLINONE", true)
	p.Profile.Client.LogExpire = util.Dur("APOLLO_AGENT_CLIENT_LOGEXPIRE", 0)
	p.Profile.Client.Ip = util.Str("APOLLO_AGENT_CLIENT_IP", "")
	p.Profile.Client.BeatFreQ = util.Dur("APOLLO_AGENT_CLIENT_BEATFREQ", 0)

	p.Profile.Server.Cluster = strings.ToLower(util.Str("APOLLO_AGENT_SERVER_CLUSTER", ""))
	p.Profile.Server.Address = util.Str("APOLLO_AGENT_SERVER_ADDRESS", "")

	p.Profile.Apps = []*App{
		{
			AppId:      util.Str("APOLLO_AGENT_APP_ID", ""),
			Namespaces: strings.Split(util.Str("APOLLO_AGENT_APP_NAMESPACES", ""), ","),
			Secret:     util.Str("APOLLO_AGENT_APP_SECRET", ""),
			InOne: &InOne{
				FileName: util.Str("APOLLO_AGENT_APP_CONFIG_NAME", ""),
				Syntax:   util.Str("APOLLO_AGENT_APP_CONFIG_SYNTAX", ""),
			},
			PollInterval: util.Dur("APOLLO_AGENT_APP_POLL_INTERVAL", 0),
		},
	}
	if util.Str("APOLLO_AGENT_APP_ID", "") == "" {
		return fmt.Errorf("[ERROR] ENV Variable APOLLO_AGENT_APP_ID is null")
	}
	log.Println("[INFO] load boot config from system ENV variables")
	return nil
}

func (p *ProfileLauncher) loadConfigFile() error {
	if _, err := os.Stat(*p.agent.Args.ConfigFile); os.IsNotExist(err) {
		return err
	}
	configs, err := ioutil.ReadFile(*p.agent.Args.ConfigFile)
	if err != nil {
		return fmt.Errorf("[ERROR] ReadFile app config file(default is app.yaml) error, " + err.Error())
	}
	err = yaml.Unmarshal(configs, &p.Profile)
	if err != nil {
		return fmt.Errorf("[ERROR] Unmarshal config file(default is app.yaml) error, " + err.Error())
	}
	p.Profile.Server.Cluster = strings.ToLower(p.Profile.Server.Cluster)

	log.Println("[INFO] load config from " + *p.agent.Args.ConfigFile)
	return nil
}

func (p *ProfileLauncher) watchConfigFile() {
	if !p.FromEnvVar {
		for {
			select {
			case event, ok := <-p.watcher.Events:
				if !ok {
					return
				}
				log.Println("[INFO] event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("[INFO] apolloConfig restart...")
					p.ProfileUpdate = true
					p.agent.SigBus.RestartS <- struct{}{}
				}

			case err, ok := <-p.watcher.Errors:
				if !ok {
					return
				}
				log.Println("[WARNING] error:", err)
			}
		}
	}
}

func (p *Profile) wrapper() {
	if p.Client != nil {
		if p.Client.Type == "" {
			p.Client.Type = _defaultClientType
		}
		if p.Client.LogExpire == 0 {
			p.Client.LogExpire = _defaultClientLogExpire
		}
	} else {
		p.Client = &Client{
			Type:      _defaultClientType,
			AllInOne:  _defaultClientAllInOne,
			LogExpire: _defaultClientLogExpire,
		}
	}
	if p.Server != nil {
		if p.Server.Cluster == "" {
			p.Server.Cluster = _defaultServerCluster
		}
	} else {
		p.Server = &Server{
			Cluster: _defaultServerCluster,
		}
	}
	if len(p.Apps) > 0 {
		for _, app := range p.Apps {
			if len(app.Namespaces) == 0 {
				app.Namespaces = []string{_defaultAppNamespace}
			}
			if app.PollInterval == 0 {
				app.PollInterval = _defaultAppPollInterval
			}
			if app.InOne == nil {
				app.InOne = &InOne{
					FileName: "." + string(os.PathSeparator) + _defaultAppNamespace,
					Syntax:   _defaultAppSyntax,
				}
			} else {
				if app.InOne.Syntax == "" {
					app.InOne.Syntax = _defaultAppSyntax
				}
				if app.InOne.FileName == "" {
					app.InOne.FileName = "." + string(os.PathSeparator) + _defaultAppNamespace
				}
			}
		}
	} else {
		p.Apps = []*App{
			{
				Namespaces:   []string{_defaultAppNamespace},
				PollInterval: _defaultAppPollInterval,
				InOne: &InOne{
					FileName: "." + string(os.PathSeparator) + _defaultAppNamespace,
					Syntax:   _defaultAppSyntax,
				},
			},
		}
	}
}

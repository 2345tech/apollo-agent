package boot

import (
	"context"
	"github.com/2345tech/apollo-agent/common"
	"log"
	"os"
	"sync"
	"time"
)

const (
	VERSION  = "v4.0.0"
	AUTHOR   = "lixy<lixy@2345.com>"
	FilePerm = 0644
)

type AgentLauncher interface {
	Init(agent *Agent) error
	Run() error
	Shutdown()
}

type LauncherFunc func(*Agent) error

func WithLauncher(launcher AgentLauncher) LauncherFunc {
	return func(agent *Agent) error {
		agent.Launchers = append(agent.Launchers, launcher)
		return launcher.Init(agent)
	}
}

type SignalBus struct {
	StopS    chan struct{}
	RestartS chan struct{}
}

type Agent struct {
	isRunning  bool
	Args       *Args
	EnvProfile bool
	LogFile    *os.File
	LogExpire  time.Duration
	BeatFreQ   time.Duration

	LogL    *LogLauncher
	ConfigL *ProfileLauncher
	SignalL *SignalLauncher

	LFunc     []LauncherFunc
	Launchers []AgentLauncher

	Context context.Context
	Cancel  context.CancelFunc
	Wg      *sync.WaitGroup

	Handlers []common.AgentHandler

	SigBus *SignalBus
}

func New(lfs ...LauncherFunc) *Agent {
	if len(lfs) <= 0 {
		panic("[ERROR] not found LauncherFunc")
	}
	args := NewArg()
	agent := &Agent{
		Args:      args,
		LFunc:     lfs,
		Launchers: make([]AgentLauncher, 0),
		Handlers:  make([]common.AgentHandler, 0),
		SigBus: &SignalBus{
			StopS:    make(chan struct{}),
			RestartS: make(chan struct{}),
		},
	}
	args.Init(agent)

	return agent
}

func (a *Agent) Init() *Agent {
	for _, p := range a.LFunc {
		if err := p(a); err != nil {
			log.Println(err.Error())
			panic("[PANIC] agent SetUp failed")
		}
	}
	return a
}

func (a *Agent) RegisterHandler(handlers ...common.AgentHandler) {
	if len(handlers) > 0 {
		for _, handler := range handlers {
			a.Handlers = append(a.Handlers, handler)
		}
	} else {
		return
	}
	a.Handlers = handlers
}

func (a *Agent) Start() error {
	if err := a.running(); err != nil {
		return err
	}

	a.Wg = new(sync.WaitGroup)
	a.Wg.Add(1)
	go a.signalBusBooting(a.Wg)

	a.Wg.Wait()
	a.ShutDown()
	return nil
}

func (a *Agent) signalBusBooting(wg *sync.WaitGroup) {
	defer wg.Done()
	if a.BeatFreQ == 0 {
		a.BeatFreQ = 10 * time.Minute
	}
	log.Println("[INFO] signalBus boot...")
	for {
		select {
		case <-a.SigBus.StopS:
			a.Stop()
			log.Println("[INFO] agent stopped")
			return

		case <-a.SigBus.RestartS:
			a.Restart()
			log.Println("[INFO] agent restarted")

		case <-time.After(a.BeatFreQ):
			log.Println("[INFO] agent heart beating")
		}
	}
}

func (a *Agent) running() error {
	if a.isRunning {
		return nil
	}
	a.Context, a.Cancel = context.WithCancel(context.Background())
	for _, p := range a.Launchers {
		if err := p.Run(); err != nil {
			return err
		}
	}

	for _, handler := range a.Handlers {
		if err := handler.PreHandle(a.Context); err != nil {
			log.Println(err.Error())
			panic("[PANIC] agent RegisterHandler failed")
		}
	}

	handlerParam := a.fillHandlerParam()
	runMode := common.ModePoll
	if a.ConfigL.Profile.Client.Type == common.ModeWatch {
		runMode = common.ModeWatch
	}

	for _, handler := range a.Handlers {
		handler.SetRunMode(runMode)
		if err := handler.PostHandle(handlerParam, a.Context); err != nil {
			return err
		}
	}
	a.isRunning = true
	return nil
}

func (a *Agent) Restart() {
	if a.isRunning {
		a.Stop()
	}

	if err := a.running(); err != nil {
		log.Println(err.Error())
		log.Println("[PANIC] agent Restart failed")
	}
}

func (a *Agent) Stop() {
	if !a.isRunning {
		return
	}
	a.Cancel()
	hLen := len(a.Handlers)
	for i := hLen - 1; i >= 0; i-- {
		handler := a.Handlers[i]
		if err := handler.AfterCompletion(a.Context); err != nil {
			log.Println(err.Error())
		}
	}
	a.isRunning = false
}

func (a *Agent) ShutDown() {
	for _, p := range a.Launchers {
		p.Shutdown()
	}
	time.Sleep(1 * time.Second)
	log.Println("[INFO] agent shutdown")
	a.LogFile.Close()
}

func (a *Agent) fillHandlerParam() *common.HandlerParam {
	param := &common.HandlerParam{
		Address:  a.ConfigL.Profile.Server.Address,
		Cluster:  a.ConfigL.Profile.Server.Cluster,
		ClientIp: a.ConfigL.Profile.Client.Ip,
		AllInOne: a.ConfigL.Profile.Client.AllInOne,
		Apps:     make([]*common.App, 0),
	}
	for _, app := range a.ConfigL.Profile.Apps {
		param.Apps = append(param.Apps, &common.App{
			AppId:        app.AppId,
			Namespaces:   app.Namespaces,
			Secret:       app.Secret,
			PollInterval: app.PollInterval,
			FileName:     app.InOne.FileName,
			Syntax:       app.InOne.Syntax,
		})
	}
	return param
}

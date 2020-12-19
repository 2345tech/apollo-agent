package apollo

import (
	"context"
	"github.com/2345tech/apolloclient"
	"log"
	"sync"
	"time"
)

type Poll struct {
	allInOne bool
	interval time.Duration
	update   chan struct{}

	Meta *MetaConfig
	Data *sync.Map

	client *apolloclient.Client
}

func NewPoller(allInOne bool, interval time.Duration) Worker {
	return &Poll{
		allInOne: allInOne,
		interval: interval,
		update:   make(chan struct{}),
		Data:     new(sync.Map),
	}
}

func (p *Poll) SetMeta(meta *MetaConfig) {
	p.Meta = meta
}

func (p *Poll) GetConfig(wg *sync.WaitGroup, ctx context.Context) {
	if ac, err := getApolloClient(p.Meta.Address, ctx); err == nil {
		p.client = ac
	} else {
		log.Println("[ERROR] NewConfigService error " + err.Error())
		return
	}
	for _, ns := range p.Meta.Namespaces {
		wg.Add(1)
		param := apolloclient.GetConfigParam{
			AppID:     p.Meta.AppId,
			Cluster:   p.Meta.Cluster,
			Namespace: ns,
			Secret:    p.Meta.Secret,
			ClientIP:  p.Meta.ClientIp,
		}
		go p.polling(param, wg, ctx)
	}
}

func (p *Poll) GetMeta() *MetaConfig {
	return p.Meta
}

func (p *Poll) GetChan() chan struct{} {
	return p.update
}

func (p *Poll) CloseChan() {
	close(p.update)
}

func (p *Poll) IsAllInOne() bool {
	return p.allInOne
}

func (p *Poll) GetData() *sync.Map {
	return p.Data
}

func (p *Poll) CleanData() {
	p.Data = new(sync.Map)
}

func (p *Poll) DeleteDataKey(key string) {
	p.Data.Delete(key)
}

func (p *Poll) polling(param apolloclient.GetConfigParam, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] [appId] %v [Namespace] %v down...\n", param.AppID, param.Namespace)
			return
		default:
			log.Printf("[INFO] [appId] %v [Namespace] %v polling...\n", param.AppID, param.Namespace)
			if data, err := p.client.GetConfig(&param); err == nil {
				if len(data.Configs) > 0 {
					p.Data.Store(param.Namespace, data.Configs)
					p.update <- struct{}{}
				}
			} else {
				log.Println("[ERROR] GetConfig from Apollo Config Service error:" + err.Error())
			}
			time.Sleep(p.interval)
		}
	}
}

package apollo

import (
	"context"
	"github.com/2345tech/apolloclient"
	"log"
	"sync"
	"time"
)

const (
	modePoll  = "poll"
	modeWatch = "watch"
)

type DefaultWorker struct {
	mode     string
	allInOne bool
	interval time.Duration
	update   chan struct{}

	Meta *MetaConfig
	Data *sync.Map

	client *apolloclient.Client
}

func NewDefaultWorker(allInOne bool, interval time.Duration, mode string) WorkerContract {
	return &DefaultWorker{
		mode:     mode,
		allInOne: allInOne,
		interval: interval,
		update:   make(chan struct{}),
		Data:     new(sync.Map),
	}
}

func (w *DefaultWorker) SetMeta(meta *MetaConfig) {
	w.Meta = meta
}

func (w *DefaultWorker) GetConfig(wg *sync.WaitGroup, ctx context.Context) {
	if ac, err := getApolloClient(w.Meta.Address, ctx); err == nil {
		w.client = ac
	} else {
		log.Println("[ERROR] NewConfigService error " + err.Error())
		return
	}
	for _, ns := range w.Meta.Namespaces {
		wg.Add(1)
		param := apolloclient.GetConfigParam{
			AppID:     w.Meta.AppId,
			Cluster:   w.Meta.Cluster,
			Namespace: ns,
			Secret:    w.Meta.Secret,
			ClientIP:  w.Meta.ClientIp,
		}

		switch w.mode {
		case modePoll:
			go w.polling(param, wg, ctx)
		case modeWatch:
			go w.watching(param, wg, ctx)
		}
	}
}

func (w *DefaultWorker) GetMeta() *MetaConfig {
	return w.Meta
}

func (w *DefaultWorker) GetChan() chan struct{} {
	return w.update
}

func (w *DefaultWorker) CloseChan() {
	close(w.update)
}

func (w *DefaultWorker) IsAllInOne() bool {
	return w.allInOne
}

func (w *DefaultWorker) GetData() *sync.Map {
	return w.Data
}

func (w *DefaultWorker) CleanData() {
	w.Data = new(sync.Map)
}

func (w *DefaultWorker) DeleteDataKey(key string) {
	w.Data.Delete(key)
}

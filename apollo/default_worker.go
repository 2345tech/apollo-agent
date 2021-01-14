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

func (w *DefaultWorker) polling(param apolloclient.GetConfigParam, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] [appId] %v [Namespace] %v down...\n", param.AppID, param.Namespace)
			return
		default:
			log.Printf("[INFO] [appId] %v [Namespace] %v polling...\n", param.AppID, param.Namespace)
			if data, err := w.client.GetConfig(&param); err == nil {
				if len(data.Configs) > 0 {
					w.Data.Store(param.Namespace, data.Configs)
					w.update <- struct{}{}
				}
			} else {
				log.Println("[ERROR] GetConfig from Apollo Config Service error:" + err.Error())
			}
			time.Sleep(w.interval)
		}
	}
}


func (w *DefaultWorker) watching(param apolloclient.GetConfigParam, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	notificationParam := &apolloclient.GetNotificationsParam{
		AppID:         param.AppID,
		Cluster:       param.Cluster,
		Secret:        param.Secret,
		Notifications: make([]apolloclient.Notification, 0),
	}
	notificationParam.Notifications = []apolloclient.Notification{
		apolloclient.Notification{
			Namespace:      param.Namespace,
			NotificationID: 0,
		},
	}
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] [appId] %v [Namespace] %v down...\n", param.AppID, param.Namespace)
			return
		default:
			log.Printf("[INFO] [appId] %v [Namespace] %v watching...\n", param.AppID, param.Namespace)
			if update, notifications, err := w.client.GetNotifications(notificationParam); err != nil {
				log.Printf("[ERROR] [appId] %v [Namespace] %v GetNotifications from Apollo Config Service error:%v\n",
					param.AppID, param.Namespace, err.Error())
				time.Sleep(w.interval)
			} else {
				if update && len(notifications) == 1 {
					notificationParam.Notifications[0].NotificationID = notifications[0].NotificationID
					if data, err := w.client.GetConfig(&param); err == nil {
						if len(data.Configs) > 0 {
							w.Data.Store(param.Namespace, data.Configs)
							w.update <- struct{}{}
						}
						param.ReleaseKey = data.ReleaseKey
					} else {
						log.Println("[INFO] GetConfig from Apollo Config Service error:" + err.Error())
					}
				} else {
					log.Printf("[ERROR] [appId] %v [Namespace] %v GetNotifications failed...\n", param.AppID, param.Namespace)
					time.Sleep(w.interval)
				}
			}
		}
	}
}
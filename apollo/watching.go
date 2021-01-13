package apollo

import (
	"context"
	"github.com/2345tech/apolloclient"
	"log"
	"sync"
	"time"
)

type Watch struct {
	allInOne bool
	interval time.Duration
	update   chan struct{}

	Meta *MetaConfig
	Data *sync.Map

	client *apolloclient.Client
}

func NewWatcher(allInOne bool, interval time.Duration) Worker {
	return &Watch{
		allInOne: allInOne,
		interval: interval,
		update:   make(chan struct{}),
		Data:     new(sync.Map),
	}
}

func (w *Watch) SetMeta(meta *MetaConfig) {
	w.Meta = meta
}

func (w *Watch) GetConfig(wg *sync.WaitGroup, ctx context.Context) {
	if ac, err := getApolloClient(w.Meta.Address, ctx); err == nil {
		w.client = ac
	} else {
		log.Println("[ERROR] NewConfigService error " + err.Error())
		return
	}
	for _, ns := range w.Meta.Namespaces {
		wg.Add(1)
		go w.watching(apolloclient.GetConfigParam{
			AppID:     w.Meta.AppId,
			Cluster:   w.Meta.Cluster,
			Namespace: ns,
			Secret:    w.Meta.Secret,
			ClientIP:  w.Meta.ClientIp,
		}, wg, ctx)
	}
}

func (w *Watch) GetMeta() *MetaConfig {
	return w.Meta
}

func (w *Watch) GetChan() chan struct{} {
	return w.update
}

func (w *Watch) CloseChan() {
	close(w.update)
}

func (w *Watch) IsAllInOne() bool {
	return w.allInOne
}

func (w *Watch) GetData() *sync.Map {
	return w.Data
}

func (w *Watch) CleanData() {
	w.Data = new(sync.Map)
}

func (w *Watch) DeleteDataKey(key string) {
	w.Data.Delete(key)
}

func (w *Watch) watching(param apolloclient.GetConfigParam, wg *sync.WaitGroup, ctx context.Context) {
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
					param.AppID, param.Namespace,err.Error())
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

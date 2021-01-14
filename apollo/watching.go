package apollo

import (
	"context"
	"github.com/2345tech/apolloclient"
	"log"
	"sync"
	"time"
)

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

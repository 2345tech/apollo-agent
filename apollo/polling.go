package apollo

import (
	"context"
	"github.com/2345tech/apolloclient"
	"log"
	"sync"
	"time"
)

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

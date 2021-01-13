package apollo

import (
	"context"
	"github.com/2345tech/apollo-agent/common"
	"github.com/2345tech/apollo-agent/util"
	"github.com/2345tech/apolloclient"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const TmpFileSuffix = ".tmp"

type Worker interface {
	SetMeta(meta *MetaConfig)
	GetMeta() *MetaConfig
	GetConfig(wg *sync.WaitGroup, ctx context.Context)
	GetChan() chan struct{}
	CloseChan()
	CleanData()
	GetData() *sync.Map
	DeleteDataKey(key string)
	IsAllInOne() bool
}

type MetaConfig struct {
	Address    string
	Cluster    string
	ClientIp   string
	AppId      string
	Secret     string
	Namespaces []string
	FileName   string
	Syntax     string
}

type ConfigData map[string]map[string]string

type Apollo struct {
	runMode string
	Worker  []Worker
	Wg      *sync.WaitGroup
}

func NewHandler() common.AgentHandler {
	return &Apollo{
		Worker: make([]Worker, 0),
		Wg:     new(sync.WaitGroup),
	}
}

func (a *Apollo) PreHandle(ctx context.Context) error {
	return nil
}

func (a *Apollo) SetRunMode(mode string) {
	a.runMode = mode
}

func (a *Apollo) PostHandle(param *common.HandlerParam, ctx context.Context) error {
	a.setWorkers(param)

	// Get Config Data from Apollo Config Service
	for _, worker := range a.Worker {
		worker.GetConfig(a.Wg, ctx)
	}

	// Collect Config Data Write To File
	for _, worker := range a.Worker {
		a.Wg.Add(1)
		go a.WriteData(worker, ctx)
	}

	log.Println("[INFO] apollo.Apollo handler running")
	return nil
}

func (a *Apollo) AfterCompletion(ctx context.Context) error {
	a.Wg.Wait()
	for _, worker := range a.Worker {
		worker.CloseChan()
	}
	a.Worker = make([]Worker, 0)
	log.Println("[INFO] apollo.Apollo handler stopped")
	return nil
}

func (a *Apollo) WriteData(worker Worker, ctx context.Context) {
	defer a.Wg.Done()
	for {
		meta := worker.GetMeta()
		select {
		case <-ctx.Done():
			log.Printf("[INFO] [appId] %v WriteData down...\n", meta.AppId)
			return
		case <-worker.GetChan():
			if worker.IsAllInOne() {
				if len(meta.Namespaces) == getSyncMapLen(worker.GetData()) {
					writeConfigInOneFile(meta, worker)
				}
			} else {
				writeConfigOneByOne(meta, worker)
			}
		}
	}
}

func (a *Apollo) setWorkers(param *common.HandlerParam) {
	for _, app := range param.Apps {
		worker := a.newWorker(param, app)
		worker.SetMeta(&MetaConfig{
			Address:    param.Address,
			Cluster:    param.Cluster,
			ClientIp:   param.ClientIp,
			AppId:      app.AppId,
			Secret:     app.Secret,
			Namespaces: app.Namespaces,
			FileName:   app.FileName,
			Syntax:     app.Syntax,
		})
		a.Worker = append(a.Worker, worker)
	}
}

func (a *Apollo) newWorker(param *common.HandlerParam, app *common.App) Worker {
	switch a.runMode {
	case common.ModeWatch:
		return NewWatcher(param.AllInOne, app.PollInterval)
	default:
		return NewPoller(param.AllInOne, app.PollInterval)
	}
}

func getApolloClient(address string, ctx context.Context) (*apolloclient.Client, error) {
	var err error
	var client *apolloclient.Client
	var body io.Reader
	var request *http.Request
	if client, err = apolloclient.NewClient(address, http.DefaultClient, nil); err != nil {
		return nil, err
	}
	if request, err = http.NewRequestWithContext(ctx, http.MethodGet, "", body); err == nil {
		client.Request = request
	}
	return client, nil
}

func writeConfigInOneFile(meta *MetaConfig, worker Worker) {
	tmpFile := meta.FileName + TmpFileSuffix
	if err := util.MultiNSInOneFile(tmpFile, meta.Syntax, meta.Namespaces, getSyncMapData(worker.GetData()));
		err != nil {
		log.Printf("[WARN] [appId] %v WriteData error : %v \n", meta.AppId, err.Error())
	} else {
		worker.CleanData()
		if covered, err := fileCompareAndCover(tmpFile, meta.FileName); err != nil {
			log.Printf("[WARNING] copy tmp to config failed. ERR# %s \n", err.Error())
		} else if covered {
			log.Printf("[INFO] =========================NEW CONFIG SUCCESS===========================")
			log.Printf("[INFO] get a new config file. %s", meta.FileName)
		}
	}
}

func writeConfigOneByOne(meta *MetaConfig, worker Worker) {
	for ns, data := range getSyncMapData(worker.GetData()) {
		oldFile := filepath.Dir(meta.FileName) + string(os.PathSeparator) + ns
		tmpFile := oldFile + TmpFileSuffix
		if err := util.SingleNSInOneFile(tmpFile, util.NSSyntax(ns), data); err != nil {
			log.Printf("[WARN] [appId] %v [Namespace] %v WriteData error : %v \n", meta.AppId, ns, err.Error())
			continue
		}
		worker.DeleteDataKey(ns)
		if covered, err := fileCompareAndCover(tmpFile, oldFile); err != nil {
			log.Printf("[WARNING] copy tmp to config failed. ERR# %s \n", err.Error())
		} else if covered {
			log.Printf("[INFO] =========================NEW CONFIG SUCCESS===========================")
			log.Printf("[INFO] get a new config file. %s", oldFile)
		}
	}
}

func fileCompareAndCover(tmpFile, oldFile string) (bool, error) {
	md5Old, _ := util.HashFileMd5(oldFile)
	md5Tmp, _ := util.HashFileMd5(tmpFile)
	if md5Old != md5Tmp {
		return true, util.CopyFile(tmpFile, oldFile)
	} else {
		return false, nil
	}
}

func getSyncMapData(syncMap *sync.Map) ConfigData {
	dataMap := make(ConfigData)
	syncMap.Range(func(ns, data interface{}) bool {
		dataMap[ns.(string)] = data.(map[string]string)
		return true
	})
	return dataMap
}

func getSyncMapLen(syncMap *sync.Map) int {
	length := 0
	syncMap.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

package boot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogLauncher struct {
	booted    bool
	agent     *Agent
	stopLog   chan bool
	LogExpire time.Duration
	Logfile   *os.File
}

func NewLog() *LogLauncher {
	return &LogLauncher{
		booted:  false,
		stopLog: make(chan bool),
	}
}

func (l *LogLauncher) Init(agent *Agent) error {
	l.agent = agent
	agent.LogL = l
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.OpenFile(*l.agent.Args.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, FilePerm)
	if err != nil {
		return fmt.Errorf("[ERROR] set log file failed! err : %v", err)
	}
	log.SetOutput(f)
	l.Logfile = f
	l.LogExpire = agent.LogExpire
	agent.LogFile = f
	return nil
}

func (l *LogLauncher) Run() error {
	if l.agent.EnvProfile {
		return nil
	}
	if l.booted {
		return nil
	}
	go l.logrotate()
	l.booted = true
	return nil
}

func (l *LogLauncher) Stop() {
}

func (l *LogLauncher) Shutdown() {
	if l.agent.EnvProfile {
		l.stopLog <- true
	}
	close(l.stopLog)
	l.booted = false
	log.Println("[INFO] LogLauncher stopped")
}

func (l *LogLauncher) logrotate() {
	lastSplitDay := time.Now().Day()
	for {
		select {
		case <-l.stopLog:
			log.Println("[INFO] LogLauncher.splitLog closed")
			return

		case <-time.After(600 * time.Second): // 10分钟check一次，模拟一个简单的cron
			now := time.Now()
			if lastSplitDay == now.Day() {
				continue
			}
			lastSplitDay = now.Day()

			logFileName := l.agent.Args.LogFile
			bakFileName := fmt.Sprintf("%s_%s", *logFileName, now.AddDate(0, 0, -1).Format("20060102"))
			_ = l.Logfile.Close()

			if err := os.Rename(*logFileName, bakFileName); err == nil {
				logFile, err := os.OpenFile(*logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, FilePerm)
				if err != nil {
					log.Fatalf("[ERROR] set new log file failed! err : %v", err)
				}
				log.SetOutput(logFile)
				l.Logfile = logFile
			}
			l.emptyTrash(filepath.Dir(l.Logfile.Name()), filepath.Base(l.Logfile.Name()))
		}
	}
}

func (l *LogLauncher) emptyTrash(path, trashFile string) {
	nowTime := time.Now().Unix()
	_ = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if strings.Contains(p, path+string(os.PathSeparator)+trashFile+"_") {
			if float64(nowTime-f.ModTime().Unix()) > l.LogExpire.Seconds() {
				_ = os.RemoveAll(p)
				log.Println("[INFO] remove log file:" + p)
			}
		}
		return nil
	})
}

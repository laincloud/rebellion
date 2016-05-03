package core

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/laincloud/lainlet/client"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

// LainAppConfHandler handles log files of lain applications
type LainAppConfHandler struct {
	lainAppConf map[string]*LainAppConfig
}

type AppInfo struct {
	PodInfos []PodInfo `json:"PodInfos"`
}

type PodInfo struct {
	InstanceNo int    `json:"InstanceNo"`
	Annotation string `json:"Annotation"`
}

type LainAppConfig struct {
	NodeName string
	AppName  string
	Procs    map[string]ProcConfig
}

type ProcConfig struct {
	ProcName     string
	InstanceNo   int
	LogDirectory string
	LogFile      string
	FilePattern  string
	Topic        string
}

// NewLainAppConfHandler initializes a pointer of LainAppConfHandler
func NewLainAppConfHandler() *LainAppConfHandler {
	return &LainAppConfHandler{
		lainAppConf: make(map[string]*LainAppConfig),
	}
}

// DynamicallyHandle implements DynamicalHandler
func (lh *LainAppConfHandler) DynamicallyHandle(update chan int) {
	lainletCli := client.New(lainletURL)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	var isChanged bool
	for {
		time.Sleep(3 * time.Second)
		ch, err := lainletCli.Watch("/v2/rebellion/localprocs", ctx)
		if err != nil {
			log.Warn("Connect to lainlet failed. Reconnect in 10 seconds")
			continue
		}
		for resp := range ch {
			time.Sleep(3 * time.Second)
			if resp.Event == "init" || resp.Event == "update" || resp.Event == "delete" {
				data := make(map[string]AppInfo)
				if err := json.Unmarshal(resp.Data, &data); err != nil {
					log.Errorf("Unmarshal error: %s", err.Error())
					continue
				}
				isChanged = false
				newLogSet := make(map[string]*LainAppConfig)
				for procName, appInfo := range data {
					appName := strings.Split(procName, ".")[0]
					prefix := filepath.Join(volumeDir, appName, procName)
					for _, podInfo := range appInfo.PodInfos {
						var annotation struct {
							Logs []string `json:"logs"`
						}
						if err := json.Unmarshal([]byte(podInfo.Annotation), &annotation); err != nil {
							log.Errorf("Unmarshal logs in annotation error: %s\n", err.Error())
							continue
						}
						for _, logPath := range annotation.Logs {
							if _, exist := newLogSet[appName]; !exist {
								newLogSet[appName] = &LainAppConfig{
									NodeName: hostName,
									AppName:  appName,
									Procs:    make(map[string]ProcConfig),
								}
							}
							currentAppConf := newLogSet[appName]
							fullPath := filepath.Clean(filepath.Join(prefix, strconv.Itoa(podInfo.InstanceNo), containerLogsDir, logPath))
							dir, filename := filepath.Split(fullPath)
							currentAppConf.Procs[fmt.Sprintf("%x", md5.Sum([]byte(fullPath)))] = ProcConfig{
								ProcName:     procName,
								InstanceNo:   podInfo.InstanceNo,
								LogDirectory: dir,
								FilePattern:  wildcardToRegex(filename),
								LogFile:      filename,
								Topic:        strings.Join([]string{procName, "log", logPath}, "."),
							}
						}
					}
				}

				//Compare and update app config
				for appName, appConfig := range lh.lainAppConf {
					fileName := fmt.Sprintf("lainapp_%s.toml", appName)
					//Clear old app log config
					if newAppConfig, exist := newLogSet[appName]; !exist {
						isChanged = true
						delete(lh.lainAppConf, appName)
						if err := os.Remove(filepath.Join(confDir, fileName)); err != nil {
							log.Errorf("Remove %s failed: %s", fileName, err.Error())
						} else {
							log.Debugf("Remove %s successfully!", fileName)
						}
					} else if !reflect.DeepEqual(*newAppConfig, *appConfig) {
						// Update existing app log config
						isChanged = true
						lh.lainAppConf[appName] = newAppConfig
						renderTemplate(lainAppTomlTmpl, fileName, newAppConfig)
					}
				}

				for newAppName, newAppConfig := range newLogSet {
					//Add new app log config
					if _, exist := lh.lainAppConf[newAppName]; !exist {
						isChanged = true
						lh.lainAppConf[newAppName] = newAppConfig
						renderTemplate(lainAppTomlTmpl, fmt.Sprintf("lainapp_%s.toml", newAppName), newAppConfig)
					}
				}

				if isChanged {
					update <- 1
				}
			} else if resp.Event != "heartbeat" {
				log.Errorf("Get lainlet event error: %s", string(resp.Data))
				break
			} else {
				log.Debugf("Get skipped event %s", resp.Event)
			}
		}
	}
}

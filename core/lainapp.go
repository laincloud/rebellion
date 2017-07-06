package core

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/laincloud/lainlet/client"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

// LainAppConfHandler handles log files of lain applications
type LainAppConfHandler struct {
	logInfos []LogInfo
}

type AppInfo struct {
	PodInfos []PodInfo `json:"PodInfos"`
}

type PodInfo struct {
	InstanceNo int    `json:"InstanceNo"`
	AppVersion string `json:"AppVersion"`
	Annotation string `json:"Annotation"`
}

type LogInfo struct {
	AppName    string
	AppVersion string
	ProcName   string
	InstanceNo int
	LogFile    string
}

func NewLainAppConfHandler() *LainAppConfHandler {
	return &LainAppConfHandler{
		logInfos: make([]LogInfo, 0),
	}
}

func (lh *LainAppConfHandler) DynamicallyHandle(update chan interface{}) {
	lainletCli := client.New(lainletURL)
	for {
		time.Sleep(3 * time.Second)
		ch, err := lainletCli.Watch("/v2/rebellion/localprocs", context.Background())
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
				var newLogSet []LogInfo
				for procName, appInfo := range data {
					procParts := strings.Split(procName, ".")
					if len(procParts) == 0 {
						continue
					}
					for _, podInfo := range appInfo.PodInfos {
						var annotation struct {
							Logs []string `json:"logs"`
						}
						if err := json.Unmarshal([]byte(podInfo.Annotation), &annotation); err != nil {
							log.Errorf("Unmarshal logs in annotation error: %s\n", err.Error())
							continue
						}
						if annotation.Logs == nil {
							continue
						}
						for _, logName := range annotation.Logs {
							logInfo := LogInfo{
								AppName:    procParts[0],
								AppVersion: podInfo.AppVersion,
								ProcName:   procName,
								InstanceNo: podInfo.InstanceNo,
								LogFile:    logName,
							}
							newLogSet = append(newLogSet, logInfo)
						}
					}
				}
				sort.Slice(newLogSet, func(i, j int) bool {
					if newLogSet[i].ProcName != newLogSet[j].ProcName {
						return newLogSet[i].ProcName < newLogSet[j].ProcName
					} else if newLogSet[i].InstanceNo != newLogSet[j].InstanceNo {
						return newLogSet[i].InstanceNo < newLogSet[j].InstanceNo
					}
					return newLogSet[i].LogFile < newLogSet[j].LogFile
				})
				if !reflect.DeepEqual(newLogSet, lh.logInfos) {
					dumpData, _ := json.Marshal(newLogSet)
					log.Infof("LogInfo is updated: %s", dumpData)
					lh.logInfos = newLogSet
					update <- lh.logInfos
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

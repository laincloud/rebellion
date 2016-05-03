package core

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/client"
)

// WebrouterConfHandler renders webrouter.toml dynamically
type WebrouterConfHandler struct {
	Containers map[string]NodeContainerInfo
}

// NodeContainerInfo represents the webouter procname and instanceNo
type NodeContainerInfo struct {
	ProcName   string `json:"proc"`
	InstanceNo int    `json:"instanceNo"`
}

// NewWebrouterConfHandler initalizes a pointer of WebrouterConfHandler
func NewWebrouterConfHandler() *WebrouterConfHandler {
	return &WebrouterConfHandler{}
}

// DynamicallyHandle implements DynamicalHandler
func (wh *WebrouterConfHandler) DynamicallyHandle(update chan int) {
	lainletCli := client.New(lainletURL)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	for {
		time.Sleep(3 * time.Second)
		ch, err := lainletCli.Watch("/v2/containers?nodename="+hostName, ctx)
		if err != nil {
			log.Warn("Connect to lainlet failed. Reconnect in 10 seconds")
			continue
		}
		for resp := range ch {
			time.Sleep(3 * time.Second)
			if resp.Event == "init" || resp.Event == "update" || resp.Event == "delete" {
				tmpMap := make(map[string]NodeContainerInfo)
				if err := json.Unmarshal(resp.Data, &tmpMap); err != nil {
					log.Errorf("Unmarshal error: %s", err.Error())
					continue
				}
				for id, container := range tmpMap {
					if container.ProcName != "webrouter.worker.worker" {
						delete(tmpMap, id)
					}
				}

				if !reflect.DeepEqual(tmpMap, wh.Containers) {
					wh.Containers = tmpMap
					renderTemplate("webrouter.toml.tmpl", "webrouter.toml", wh)
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

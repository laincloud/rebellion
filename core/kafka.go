package core

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/client"
)

// KafkaConfHandler renders kafka.toml dynamically
type KafkaConfHandler struct {
	KafkaAddrs string
}

// NewKafkaConfHandler initalizes a pointer of KafkaConfHandler with empty KafkaAddrs
func NewKafkaConfHandler() *KafkaConfHandler {
	return &KafkaConfHandler{}
}

// DynamicallyHandle implements DynamicalHandler
func (kc *KafkaConfHandler) DynamicallyHandle(update chan int) {
	lainletCli := client.New(lainletURL)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	for {
		time.Sleep(3 * time.Second)
		ch, err := lainletCli.Watch("/v2/configwatcher?target=kafka", ctx)
		if err != nil {
			log.Warn("Connect to lainlet failed. Reconnect in 10 seconds")
			continue
		}
		for resp := range ch {
			time.Sleep(3 * time.Second)
			if resp.Event == "init" || resp.Event == "update" || resp.Event == "delete" {
				dataMap := make(map[string]string)
				if err := json.Unmarshal(resp.Data, &dataMap); err != nil {
					log.Errorf("Unmarshal error: %s", err.Error())
					continue
				}
				var data []string
				if kafkaConf, exist := dataMap["kafka"]; exist {
					if err := json.Unmarshal([]byte(kafkaConf), &data); err != nil {
						log.Errorf("Unmarshal error: %s", err.Error())
						continue
					}
					sort.Strings(data)
				} else {
					data = make([]string, 0)
				}
				newKafkaAddrBytes, _ := json.Marshal(data)
				newKafkaAddrs := string(newKafkaAddrBytes)
				if newKafkaAddrs != kc.KafkaAddrs {
					kc.KafkaAddrs = newKafkaAddrs
					renderTemplate("kafka.toml.tmpl", "kafka.toml", kc)
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

package core

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/laincloud/lainlet/client"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

// KafkaAddressList represents the json string of Kafka addresses list
type KafkaAddressList string

// KafkaConfHandler renders kafka.toml dynamically
type KafkaConfHandler struct {
	KafkaAddrs KafkaAddressList
}

// NewKafkaConfHandler initalizes a pointer of KafkaConfHandler with empty KafkaAddrs
func NewKafkaConfHandler() *KafkaConfHandler {
	return &KafkaConfHandler{}
}

// DynamicallyHandle implements DynamicalHandler
func (kc *KafkaConfHandler) DynamicallyHandle(update chan interface{}) {
	lainletCli := client.New(lainletURL)
	for {
		time.Sleep(3 * time.Second)
		ch, err := lainletCli.Watch("/v2/configwatcher?target=kafka", context.Background())
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
				newKafkaAddrs := KafkaAddressList(newKafkaAddrBytes)
				if newKafkaAddrs != kc.KafkaAddrs {
					log.Infof("Kafka address is updated: %s", newKafkaAddrs)
					kc.KafkaAddrs = newKafkaAddrs
					update <- kc.KafkaAddrs
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

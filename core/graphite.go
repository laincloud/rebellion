package core

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/mijia/sweb/log"
	"github.com/laincloud/lainlet/client"
)

const defaultGraphitePort = "2003"

// GraphiteConfHandler renders graphite.toml once.
// GraphiteConfHandler will get GraphiteVIP from lainlet when calling StaticallyHandle
type GraphiteConfHandler struct {
	GraphitePort string
}

type VIPConfig struct {
	AppName  string `json:"app"`
	ProcName string `json:"proc"`
	Ports    []PortConfig
}

type PortConfig struct {
	Destination string `json:"dest"`
	Source      string `json:"src"`
	Protocol    string `json:"proto"`
}

// NewGraphiteConfHandler initalizes a pointer of KafkaConfHandler with empty GraphiteVIP
func NewGraphiteConfHandler() *GraphiteConfHandler {
	return &GraphiteConfHandler{
		GraphitePort: getEnvWithDefault("GRAPHITE_PORT", defaultGraphitePort),
	}
}

// StaticallyHandle implements StaticalHandler
func (gc *GraphiteConfHandler) StaticallyHandle() {
	lainletCli := client.New(lainletURL)
	var config map[string]string
	if data, err := lainletCli.Get("/v2/configwatcher?target=features/graphite", time.Second); err != nil {
		log.Errorf("Get features/graphite error: %s", err.Error())
	} else if err := json.Unmarshal(data, &config); err != nil {
		log.Errorf("GraphiteConfHandler unmarshalling json error: %s", err.Error())
	} else if isGraphiteEnabled, _ := strconv.ParseBool(config["features/graphite"]); isGraphiteEnabled {
		renderTemplate("graphite.toml.tmpl", "graphite.toml", gc)
	}
}

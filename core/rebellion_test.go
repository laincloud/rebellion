package core

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderFile(t *testing.T) {
	data, _ := json.Marshal([]string{"192.168.0.11:7001", "192.168.0.12:7001"})
	testApp := Rebellion{
		LogInfos: []LogInfo{
			{
				AppName:    "test_app_1",
				ProcName:   "test_app_1.web.web",
				InstanceNo: 1,
				LogFile:    "debug.log",
			},
			{
				AppName:    "test_app_1",
				ProcName:   "test_app_1.web.web",
				InstanceNo: 1,
				LogFile:    "warn.log",
			},
		},
		LainletPort: "9001",
		KafkaAddr:   KafkaAddressList(data),
	}
	renderTemplate("../templates/filebeat.yml.tmpl", "../templates/filebeat.yml.actual", testApp)
	var expectData, actualData []byte
	var err error
	expectData, err = ioutil.ReadFile("../templates/filebeat.yml.expect")
	assert.NoError(t, err)
	actualData, err = ioutil.ReadFile("../templates/filebeat.yml.actual")
	assert.NoError(t, err)
	assert.Equal(t, string(expectData), string(actualData))
}

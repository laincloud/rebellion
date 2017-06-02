package core

import (
	"os"
	"os/exec"
	"text/template"

	"github.com/mijia/sweb/log"
)

const (
	defaultLainletPort = "9001"
	filebeatYmlTmpl    = "/etc/filebeat/filebeat.yml.tmpl"
	filebeatYml        = "/etc/filebeat/filebeat.yml"
)

// Rebellion is the data holder rendering Filebeat template
type Rebellion struct {
	update    chan interface{}
	KafkaAddr KafkaAddressList
	LogInfos  []LogInfo
}

var lainletURL = "lainlet.lain:" + getEnvWithDefault("LAINLET_PORT", defaultLainletPort)

// ListenAndUpdate is the main working goroutine of rebellion
func Run() {
	r := Rebellion{
		update: make(chan interface{}),
	}
	go NewKafkaConfHandler().DynamicallyHandle(r.update)
	go NewLainAppConfHandler().DynamicallyHandle(r.update)
	for data := range r.update {
		switch data.(type) {
		case KafkaAddressList:
			r.KafkaAddr = data.(KafkaAddressList)
		case []LogInfo:
			r.LogInfos = data.([]LogInfo)
		default:
			log.Errorf("Unknown dataType")
			continue
		}
		renderTemplate(filebeatYml, filebeatYmlTmpl, r)
		reload()
	}
}

// renderTemplate renders data to the template with templateName, and output a file with filename
func renderTemplate(templateName, targetName string, data interface{}) {
	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		log.Errorf("Parsing %s error: %s", templateName, err.Error())
		return
	}
	if confFile, err := os.Create(targetName); err != nil {
		log.Errorf("%s created failed: %s", targetName, err.Error())
	} else {
		defer confFile.Close()
		if err := tmpl.Execute(confFile, data); err != nil {
			log.Errorf("%s updated failed: %s", targetName, err.Error())
		} else {
			log.Debugf("%s updated successfully!", targetName)
		}
	}
}

// reload calls the commandline interface of supervisor to restart Filebeat
func reload() {
	if err := exec.Command("supervisorctl", "restart", "filebeat").Run(); err != nil {
		log.Errorf("Reolad filebeat failed: %s", err.Error())
	}
}

// getEnvWithDefault gets the env value of the key. If the key is not set,
// defaultVal will be returned
func getEnvWithDefault(key, defaultVal string) string {
	var val string
	if val = os.Getenv(key); val == "" {
		val = defaultVal
	}
	return val
}

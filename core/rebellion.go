package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/mijia/sweb/log"
)

// Rebellion calls staticalHandlers first, restarts Hekad, and listens the
// events of all dynamicHandlers
type Rebellion struct {
	update          chan int
	dynamicHandlers []DynamicHandler
	staticHandlers  []StaticHandler
}

// StaticHandler handles the templates rendering only once
type StaticHandler interface {
	StaticallyHandle()
}

// DynamicHandler handles the templates rendering as necessary
// Implements should pass any int value to the channel to notify rebellion
// to restart Hekad
type DynamicHandler interface {
	DynamicallyHandle(chan int)
}

const (
	volumeDir          = "/data/lain/volumes"
	defaultLainletPort = "9001"
	containerLogsDir   = "/lain/logs"
	lainAppTomlTmpl    = "lainapp.toml.tmpl"
)

var hostName, confDir, baseDir string
var lainletURL = "lainlet.lain:" + getEnvWithDefault("LAINLET_PORT", defaultLainletPort)

// NewRebellion initializes a pointer of Rebellion with all the handlers registered
func NewRebellion(host, conf, base string) *Rebellion {
	hostName, confDir, baseDir = host, conf, base
	commonConfHandler := NewCommonConfHandler()
	lainAppConfHandler := NewLainAppConfHandler()
	syslogConfHandler := NewSyslogConfHandler(host)
	graphiteConfHandler := NewGraphiteConfHandler()
	webrouterConfHandler := NewWebrouterConfHandler()
	kafkaConfHandler := NewKafkaConfHandler()

	dynamicHandlerArr := []DynamicHandler{
		lainAppConfHandler,
		kafkaConfHandler,
		webrouterConfHandler,
	}
	staticHandlerArr := []StaticHandler{
		commonConfHandler,
		syslogConfHandler,
		graphiteConfHandler,
	}

	return &Rebellion{
		update:          make(chan int),
		dynamicHandlers: dynamicHandlerArr,
		staticHandlers:  staticHandlerArr,
	}
}

// ListenAndUpdate is the main working goroutine of rebellion
func (r *Rebellion) ListenAndUpdate() {
	for _, handler := range r.staticHandlers {
		handler.StaticallyHandle()
	}
	reload()

	for _, handler := range r.dynamicHandlers {
		go handler.DynamicallyHandle(r.update)
	}

	for range r.update {
		time.Sleep(3 * time.Second)
		reload()
	}

}

// renderTemplate renders data to the template with templateName, and output a file with filename
func renderTemplate(templateName, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filepath.Join(baseDir, "templates", templateName))
	if err != nil {
		log.Errorf("Parsing %s error: %s", filename, err.Error())
		return
	}
	if confFile, err := os.Create(filepath.Join(confDir, filename)); err != nil {
		log.Errorf("%s created failed: %s", filename, err.Error())
	} else {
		defer confFile.Close()
		if err := tmpl.Execute(confFile, data); err != nil {
			log.Errorf("%s updated failed: %s", filename, err.Error())
		} else {
			log.Debugf("%s updated successfully!", filename)
		}
	}
}

// reload calls the commandline interface of supervisor to restart Hekad
func reload() {
	if err := exec.Command("supervisorctl", "-s", "unix:///var/run/supervisor.sock", "restart", "hekad").Run(); err != nil {
		log.Errorf("Reolad hekad failed: %s", err.Error())
	}
}

// wildcardToRegex converts the wildcard string to regular expression one
func wildcardToRegex(wildcard string) string {
	var result []rune
	for _, c := range wildcard {
		switch c {
		case '*':
			result = append(result, '.', '*')
		case '?':
			result = append(result, '.')
		case '(', ')', '[', ']', '$', '^', '.', '{', '}', '|', '\\', '+':
			result = append(result, '\\', rune(c))
		default:
			result = append(result, rune(c))
		}
	}
	return string(result)
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

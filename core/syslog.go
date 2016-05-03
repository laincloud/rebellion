package core

//SyslogConfHandler renders syslog_lain.toml once
type SyslogConfHandler struct {
	NodeName   string
	LainletURL string
}

//NewSyslogConfHandler initalizes a pointer of SyslogConfHandler with the given nodeName
func NewSyslogConfHandler(nodeName string) *SyslogConfHandler {
	return &SyslogConfHandler{
		NodeName:   nodeName,
		LainletURL: lainletURL,
	}
}

// StaticallyHandle implements StaticalHandler
func (sc *SyslogConfHandler) StaticallyHandle() {
	renderTemplate("syslog_lain.toml.tmpl", "syslog_lain.toml", sc)
}

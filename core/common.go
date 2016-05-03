package core

// CommonConfHandler renders common.toml with some basic configurations
type CommonConfHandler struct {
}

// NewCommonConfHandler initalizes a pointer of CommonConfHandler
func NewCommonConfHandler() *CommonConfHandler {
	return &CommonConfHandler{}
}

// StaticallyHandle implements StaticalHandler
func (cc *CommonConfHandler) StaticallyHandle() {
	renderTemplate("common.toml.tmpl", "common.toml", cc)
}

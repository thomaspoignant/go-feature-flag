package testutils

import "text/template"

func NewExportableMockEvent(name string) ExportableMockEvent {
	return ExportableMockEvent{name: name}
}

type ExportableMockEvent struct {
	name string
}

func (e ExportableMockEvent) GetUserKey() string {
	return e.name
}

func (e ExportableMockEvent) GetKey() string {
	return e.name
}

func (e ExportableMockEvent) GetCreationDate() int64 {
	return 0
}

func (e ExportableMockEvent) FormatInCSV(_ *template.Template) ([]byte, error) {
	return []byte(e.name), nil
}

func (e ExportableMockEvent) FormatInJSON() ([]byte, error) {
	return []byte(`{"name":"` + e.name + `"}`), nil
}

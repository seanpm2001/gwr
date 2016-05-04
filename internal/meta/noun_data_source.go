package meta

import (
	"text/template"

	"github.com/uber-go/gwr/source"
)

const metaNounName = "/meta/nouns"

var nounsTextTemplate = template.Must(template.New("meta_nouns_text").Parse(`
{{- define "get" -}}
{{ range $name, $info := . -}}
- {{ $name }} formats: {{ $info.Formats }}
{{ end -}}
{{- end -}}
`))

type dataSourceUpdate struct {
	Type string      `json:"type"`
	Info source.Info `json:"info"`
}

// NounDataSource provides a data source that describes other data sources.  It
// is used to implement the "/meta/nouns" data source.
type NounDataSource struct {
	sources *source.DataSources
	watcher source.GenericDataWatcher
}

// NewNounDataSource creates a new data source that gets information on other
// data sources and streams updates about them.
func NewNounDataSource(dss *source.DataSources) *NounDataSource {
	return &NounDataSource{
		sources: dss,
	}
}

// Name returns the static "/meta/nouns" string; currently using more than one
// NounDataSource in a single DataSources is unsupported.
func (nds *NounDataSource) Name() string {
	return metaNounName
}

// Attrs returns a nil descriptor to implement the GenericDataSource.
func (nds *NounDataSource) Attrs() map[string]interface{} {
	return nil
}

// TextTemplate returns a text/template to implement the GenericDataSource with
// a "text" format option.
func (nds *NounDataSource) TextTemplate() *template.Template {
	return nounsTextTemplate
}

// Get returns all currently knows data sources.
func (nds *NounDataSource) Get() interface{} {
	return nds.sources.Info()
}

// GetInit returns identical data to Get so that all Watch streams start out
// with a snapshot of the world.
func (nds *NounDataSource) GetInit() interface{} {
	return nds.Get()
}

// Watch implements GenericDataSource by retaining a reference to the passed
// watcher.  Updates are later sent to the watcher when new data sources are
// added.  Currently there is no data source removal, but when there is,
// removal updates will be sent here (TODO change this once we implement
// source removal).
func (nds *NounDataSource) Watch(watcher source.GenericDataWatcher) {
	nds.watcher = watcher
}

// SourceAdded is called whenever a source is added to the DataSources.
func (nds *NounDataSource) SourceAdded(ds source.DataSource) {
	if nds.watcher != nil {
		nds.watcher.HandleItem(dataSourceUpdate{"add", source.GetInfo(ds)})
	}
}

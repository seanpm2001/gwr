package gwr

import (
	"errors"
	"fmt"
	"io"
)

// DefaultDataSources is default data sources registry which data sources are
// added to by the module-level Add* functions.  It is used by all of the
// protocol servers if no data sources are provided.
var DefaultDataSources DataSources

func init() {
	DefaultDataSources.init()
}

// AddDataSource adds a data source to the default data sources registry.
func AddDataSource(ds DataSource) error {
	return DefaultDataSources.AddDataSource(ds)
}

// AddMarshaledDataSource adds a generically marshaled data source to the
// default data sources registry.
func AddMarshaledDataSource(gds GenericDataSource) error {
	return DefaultDataSources.AddMarshaledDataSource(gds)
}

// DataSource is the low-level interface implemented by all data sources.
//
// On formats, implementanions:
// - must implement format == "json"
// - should implement format == "text"
// - may implement any other formats that make sense for them
//
// Further implementation requirements are listed within the interface
// functions' documentation.
type DataSource interface {
	// Info returns a struct describing the formats and other capabilities of
	// this data source.  All implemented formats must be listed in
	// Info().Formats.  At least "json" should be in Info().Formats.
	Info() DataSourceInfo

	// Get implementations:
	// - may return ErrNotGetable if get is not supported by the data source
	// - if the format is not support then ErrUnsupportedFormat must be returned
	// - must format and write any available data to the supplied io.Writer
	// - should return any write error
	Get(format string, w io.Writer) error

	// Watch implementations:
	// - may return ErrNotWatchable if watch is not supported by the data
	//   source
	// - if the format is not support then ErrUnsupportedFormat must be returned
	// - may format and write initial data to the supplied io.Writer; any
	//   initial write error must be returned
	// - may retain and write to the supplied io.Writer indefinately until it
	//   returns a write error
	//
	// Note that at this level, data sources are responsible for both item
	// marshalling and stream framing.
	//
	// Framing for the required "json" format is as follows:
	// - JSON must be encoded in compact (no intermediate whitespace) form
	// - each JSON record must be separated by a newline "\n"
	//
	// Framing for the required "text" format is as follows:
	// - any initial stream data should be followed by a blank line (double new
	//   line "\n\n")
	// - items should be separated by newlines
	// - if an item's text form takes up multiple lines, it should either use
	//   indentation or a double blank line to separate itself from siblings
	Watch(format string, w io.Writer) error
}

var (
	// ErrUnsupportedFormat should be returned by DataSource.Get and
	// DataSource.Watch if the requested format is not supported.
	ErrUnsupportedFormat = errors.New("unsupported format")

	// ErrNotGetable should be returned by DataSource.Get if the data source
	// does not support get.
	ErrNotGetable = errors.New("get not supported, data source is watch-only")

	// ErrNotWatchable should be returned by DataSource.Get if the data source
	// does not support watch.
	ErrNotWatchable = errors.New("watch not supported, data source is get-only")
)

// DataSourceInfo provides a description of each
// data source, such as name and supported formats.
type DataSourceInfo struct {
	Name    string                 `json:"name"`
	Formats []string               `json:"formats"`
	Attrs   map[string]interface{} `json:"attrs"`
}

// DataSources is a flat collection of DataSources
// with a meta introspection data source.
type DataSources struct {
	sources   map[string]DataSource
	metaNouns metaNounDataSource
}

// NewDataSources creates a DataSources structure
// an sets up its "/meta/nouns" data source.
func NewDataSources() *DataSources {
	dss := &DataSources{}
	dss.init()
	return dss
}

func (dss *DataSources) init() {
	dss.sources = make(map[string]DataSource, 2)
	dss.metaNouns.sources = dss
	dss.AddMarshaledDataSource(&dss.metaNouns)
}

// Get returns the named data source or nil if none is defined.
func (dss *DataSources) Get(name string) DataSource {
	source, ok := dss.sources[name]
	if ok {
		return source
	}
	return nil
}

// Info returns a map of all DataSource.Info() data
func (dss *DataSources) Info() map[string]DataSourceInfo {
	info := make(map[string]DataSourceInfo, len(dss.sources))
	for name, ds := range dss.sources {
		info[name] = ds.Info()
	}
	return info
}

// AddMarshaledDataSource adds a generically-marshaled data source. It is a
// convenience for AddDataSource(NewMarshaledDataSource(gds, nil))
func (dss *DataSources) AddMarshaledDataSource(gds GenericDataSource) error {
	mds := NewMarshaledDataSource(gds, nil)
	// TODO: useful to return mds?
	return dss.AddDataSource(mds)
}

// AddDataSource adds a DataSource, if none is
// already defined for the given name.
func (dss *DataSources) AddDataSource(ds DataSource) error {
	info := ds.Info()
	_, ok := dss.sources[info.Name]
	if ok {
		return fmt.Errorf("data source already defined")
	}
	dss.sources[info.Name] = ds
	dss.metaNouns.dataSourceAdded(ds)
	return nil
}

// TODO: do we really need to support removing data sources?  I can see the
// case for intermediaries perhaps, and suspect that is needed... but punting
// for now.

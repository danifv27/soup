package clover

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/ostafen/clover"
)

type CloverAuditer struct {
	db         *clover.DB
	collection string
	path       string
}

func ParseURI(uri string) (string, string, error) {
	var path string
	var col string

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	if u.Scheme != "audit" {
		return "", "", fmt.Errorf("ParseURI: invalid scheme %s", u.Scheme)
	}

	switch u.Opaque {
	case "clover":
		path = u.Query().Get("path")
		if path == "" {
			return "", "", fmt.Errorf("ParseURI: path not defined")
		}
		col = u.Query().Get("collection")
		if col == "" {
			return "", "", fmt.Errorf("ParseURI: collection not defined")
		}
	default:
		return "", "", fmt.Errorf("ParseURI: unsupported audit implementation %q", u.Opaque)
	}

	return path, col, nil
}

func NewCloverAuditer(uri string) (CloverAuditer, error) {
	var d *clover.DB
	var err error
	var dbPath, col string

	if dbPath, col, err = ParseURI(uri); err != nil {
		return CloverAuditer{}, fmt.Errorf("NewCloverAuditer: %w", err)
	}
	if d, err = clover.Open(dbPath); err != nil {
		return CloverAuditer{}, fmt.Errorf("NewCloverAuditer: %w", err)
	}
	d.CreateCollection(col)

	return CloverAuditer{
		db:         d,
		collection: col,
		path:       dbPath,
	}, nil
}

func (c CloverAuditer) Log(event *audit.Event) error {
	var err error
	var data []byte
	var fields map[string]interface{}

	nowUTC := time.Now().UTC()
	event.CreatedAt = &nowUTC

	if data, err = json.Marshal(event); err != nil {
		return fmt.Errorf("Log: %w", err)
	}
	if err = json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("Log: %w", err)
	}
	// insert a new document inside the collection
	doc := clover.NewDocumentOf(fields)
	// InsertOne returns the id of the inserted document
	if _, err = c.db.InsertOne(c.collection, doc); err != nil {
		return fmt.Errorf("Log: %w", err)
	}

	return nil
}

//TODO: apply filter to query
func (c CloverAuditer) ReadLog(option *audit.ReadLogOption) ([]audit.Event, error) {
	var err error
	var docs []*clover.Document
	var size int

	if option.StartTime == nil {
		return nil, fmt.Errorf("ReadLog: %s", "start time can't be nil")
	}
	if option.EndTime == nil {
		return nil, fmt.Errorf("ReadLog: %s", "end time can't be nil")
	}
	if option.StartTime.After(*option.EndTime) {
		return nil, fmt.Errorf("ReadLog: %s", "end time can't be before start time")
	}

	query := c.db.Query(c.collection)
	if size, err = query.Count(); err != nil {
		return nil, err
	}
	//Find all documents between start and end time
	if docs, err = query.FindAll(); err != nil {
		return nil, err
	}
	events := make([]audit.Event, 0, size)
	for j := 0; j < size; j++ {
		evt := &audit.Event{}
		if err = docs[j].Unmarshal(&evt); err != nil {
			return nil, err
		}
		events = append(events, *evt)
	}

	return events, nil
}

//TODO: apply filter to query
func (c CloverAuditer) TotalCount(option *audit.ReadLogOption) (int, error) {
	var err error
	var size int

	if option.StartTime == nil {
		return -1, fmt.Errorf("TotalCount: %s", "start time can't be nil")
	}
	if option.EndTime == nil {
		return -1, fmt.Errorf("TotalCount: %s", "end time can't be nil")
	}
	if option.StartTime.After(*option.EndTime) {
		return -1, fmt.Errorf("TotalCount: %s", "end time can't be before start time")
	}

	if size, err = c.db.Query(c.collection).Count(); err != nil {
		return -1, err
	}

	return size, nil
}

// func (c CloverAuditer) Count() (int, error) {

// 	return c.db.Query(c.collection).Count()
// }

func (c *CloverAuditer) Close() error {

	return c.db.Close()
}

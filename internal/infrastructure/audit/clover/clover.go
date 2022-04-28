package clover

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/ostafen/clover"
)

type CloverAuditer struct {
	db         *clover.DB
	collection string
}

func NewCloverAuditer(dbPath string, col string) (*CloverAuditer, error) {
	var d *clover.DB
	var err error

	if d, err = clover.Open(dbPath); err != nil {
		return nil, fmt.Errorf("NewCloverAuditer: %w", err)
	}
	d.CreateCollection(col)

	return &CloverAuditer{
		db:         d,
		collection: col,
	}, nil
}

func (c *CloverAuditer) Log(event *audit.Event) error {
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

func (c *CloverAuditer) ReadLog(option *audit.ReadLogOption) ([]audit.Event, error) {
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
	for j := range events {
		evt := &audit.Event{}
		if err = docs[j].Unmarshal(&evt); err != nil {
			return nil, err
		}
		events = append(events, *evt)
	}

	return events, nil
}

func (c *CloverAuditer) Count() (int, error) {

	return c.db.Query(c.collection).Count()
}

func (c *CloverAuditer) Close() error {

	return c.db.Close()
}

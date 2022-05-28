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

func getCriteria(option *audit.GetEventOption) (*clover.Criteria, error) {

	if option.StartTime == nil {
		return nil, fmt.Errorf("getCriteria: %s", "start time can't be nil")
	}
	if option.EndTime == nil {
		now := time.Now().UTC()
		option.EndTime = &now
		// return -1, fmt.Errorf("getCriteria: %s", "end time can't be nil")
	}
	if option.StartTime.After(*option.EndTime) {
		return nil, fmt.Errorf("getCriteria: %s", "end time can't be before start time")
	}
	criteria := clover.Field("created_at").GtEq(*option.StartTime).And(clover.Field("created_at").Lt(*option.EndTime))

	return criteria, nil
}

func getQuery(auditer CloverAuditer, option *audit.GetEventOption) (*clover.Query, error) {

	query := auditer.db.Query(auditer.collection)

	if option.Limit > 0 {
		return query.Limit(option.Limit), nil
	}

	return query, nil
}

func (c CloverAuditer) Audit(event *audit.Event) error {
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
func (c CloverAuditer) GetEvents(option *audit.GetEventOption) ([]audit.Event, error) {
	var err error
	var docs []*clover.Document
	var size int
	var criteria *clover.Criteria
	var query *clover.Query

	if query, err = getQuery(c, option); err != nil {
		return nil, err
	}
	if criteria, err = getCriteria(option); err != nil {
		return nil, err
	}
	if size, err = query.Where(criteria).Count(); err != nil {
		return nil, err
	}
	//Find all documents between start and end time
	sorting := clover.SortOption{
		Field:     "created_at",
		Direction: 1,
	}
	if docs, err = query.Sort(sorting).Where(criteria).FindAll(); err != nil {
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

func (c CloverAuditer) GetNumberOfEvents(option *audit.GetEventOption) (int, error) {
	var err error
	var size int
	var query *clover.Query
	var criteria *clover.Criteria

	if option == nil {
		return c.db.Query(c.collection).Count()
	}
	if query, err = getQuery(c, option); err != nil {
		return -1, err
	}
	if criteria, err = getCriteria(option); err != nil {
		return -1, err
	}
	if size, err = query.Where(criteria).Count(); err != nil {
		return -1, err
	}

	return size, nil
}

func exportCollection(auditer CloverAuditer, col string, path string) error {

	return auditer.db.ExportCollection(col, path)
}

func (c *CloverAuditer) DropCollection(col string) error {

	return c.db.DropCollection(col)
}

func (c *CloverAuditer) HasCollection(col string) (bool, error) {

	return c.db.HasCollection(col)
}

func (c *CloverAuditer) Close() error {

	return c.db.Close()
}

func (c *CloverAuditer) UseCollection(collection string) error {
	var err error

	collectionExists, _ := c.db.HasCollection(collection)
	if !collectionExists {
		err = c.db.CreateCollection(collection)
	}
	if err == nil {
		c.collection = collection
	}

	return err
}

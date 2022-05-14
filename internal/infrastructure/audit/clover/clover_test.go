package clover

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, name string, db *CloverAuditer) (func(t *testing.T, name string, db *CloverAuditer), error) {
	log.Printf("setting up test: %s", name)
	db.collection = name
	// Check if collection already exists
	collectionExists, _ := db.HasCollection(db.collection)
	if !collectionExists {
		db.CreateCollection(db.collection)
		size, _ := db.TotalCount(nil)
		log.Printf("collection %s created, size %d", db.collection, size)
	}
	// Return a function to teardown the suite
	return func(t *testing.T, name string, db *CloverAuditer) {
		log.Printf("teardown test: %s", name)
		// Check if collection already exists
		collectionExists, _ := db.HasCollection(db.collection)
		if collectionExists {
			db.DropCollection(db.collection)
			log.Printf("collection %s dropped", db.collection)
		}
	}, nil
}

// You can use testing.T, if you want to test the code without benchmarking
func setupSuite(t *testing.T) (func(t *testing.T, db *CloverAuditer), *CloverAuditer) {

	dir, err := ioutil.TempDir("", "clover-test")
	require.NoError(t, err) //stops test execution if fail
	log.Printf("setting up suite (dir: %s)", dir)
	c, err := NewCloverAuditer(fmt.Sprintf("audit:clover?path=%s&collection=test", dir))
	require.NoError(t, err) //stops test execution if fail

	// Return a function to teardown the test
	return func(t *testing.T, db *CloverAuditer) {
		log.Println("teardown suite")
		db.Close()
	}, &c
}

func TestParseURI(t *testing.T) {
	teardownSuite, db := setupSuite(t)
	defer teardownSuite(t, db)

	type args struct {
		uri string
	}
	tests := map[string]struct {
		args       args
		beforeTest func(a *args)
		wantError  error
		wantCol    string
		wantPath   string
	}{
		"parse uri": {
			args: args{
				uri: "audit:clover?path=/tmp/soup-audit&collection=audit",
			},
			beforeTest: nil,
			wantError:  nil,
			wantCol:    "audit",
			wantPath:   "/tmp/soup-audit",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var p, c string

			teardownTest, err := setupTest(t, name, db)
			defer teardownTest(t, name, db)
			require.NoError(t, err)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}
			p, c, err = ParseURI(tt.args.uri)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("Unexpected error; got %v, want %v", err, tt.wantError)
			}
			if c != tt.wantCol {
				t.Errorf("Unexpected collection; got %v, want %v", c, tt.wantCol)
			}
			if p != tt.wantPath {
				t.Errorf("Unexpected path; got %v, want %v", c, tt.wantPath)
			}
		})
	}
}

func TestLog(t *testing.T) {
	teardownSuite, db := setupSuite(t)
	defer teardownSuite(t, db)

	type args struct {
		events []audit.Event
	}
	tests := map[string]struct {
		args       args
		beforeTest func(a *args)
		wantError  error
		wantRc     int
	}{
		"insertOneDocument": {
			args: args{},
			beforeTest: func(a *args) {
				for j := 0; j < 1; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
				}
			},
			wantError: nil,
			wantRc:    1,
		},
		"insertMultipleDocuments": {
			args: args{},
			beforeTest: func(a *args) {
				for j := 0; j < 7; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/develop-%d", j),
					})
				}
			},
			wantError: nil,
			wantRc:    7, //Because we do not remove elements from previous test
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t, name, db)
			defer teardownTest(t, name, db)
			require.NoError(t, err)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}
			for i := range tt.args.events {
				err = db.Log(&tt.args.events[i])
				if !errors.Is(err, tt.wantError) {
					t.Errorf("Unexpected error; got %v, want %v", err, tt.wantError)
				}
			}
			regs, err := db.TotalCount(nil)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("Unexpected error; got %v, want %v", err, tt.wantError)
			}
			if regs != tt.wantRc {
				t.Errorf("Wrong number of registers; got %v, want %v", regs, tt.wantRc)
			}
		})
	}
}

func TestReadLog(t *testing.T) {
	teardownSuite, db := setupSuite(t)
	defer teardownSuite(t, db)

	type args struct {
		events []audit.Event
		option *audit.ReadLogOption
	}

	tests := map[string]struct {
		args       args
		beforeTest func(a *args)
		wantError  error
		wantRc     int
	}{
		"readAllDocuments": {
			args: args{},
			beforeTest: func(a *args) {
				a.option = new(audit.ReadLogOption)
				start := time.Now().Add(-time.Minute * 5).UTC()
				a.option.StartTime = &start
				end := time.Now().Add(time.Minute * 5).UTC()
				a.option.EndTime = &end
				for j := 0; j < 8; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
					err := db.Log(&a.events[j])
					require.NoError(t, err) //stops test execution if fail
				}
				exportCollection(*db, db.collection, fmt.Sprintf("%s/%s.json", db.path, db.collection))
			},
			wantError: nil,
			wantRc:    8,
		},
		"readLimitedDocuments": {
			args: args{},
			beforeTest: func(a *args) {
				a.option = new(audit.ReadLogOption)
				start := time.Now().Add(-time.Minute * 5).UTC()
				a.option.StartTime = &start
				end := time.Now().Add(time.Minute * 5).UTC()
				a.option.EndTime = &end
				a.option.Limit = 4
				for j := 0; j < 8; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
					err := db.Log(&a.events[j])
					require.NoError(t, err) //stops test execution if fail
				}
				exportCollection(*db, db.collection, fmt.Sprintf("%s/%s.json", db.path, db.collection))
			},
			wantError: nil,
			wantRc:    4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t, name, db)
			defer teardownTest(t, name, db)
			require.NoError(t, err)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}
			events, err := db.ReadLog(tt.args.option)
			if err != nil {
				t.Errorf("Can not read audit log; got %v", err)
			}
			if len(events) != tt.wantRc {
				t.Errorf("Wrong number of registers; got %v, want %v", len(events), tt.wantRc)
			}
		})
	}
}

func TestTotalCount(t *testing.T) {
	teardownSuite, db := setupSuite(t)
	defer teardownSuite(t, db)

	type args struct {
		events []audit.Event
		option *audit.ReadLogOption
	}

	tests := map[string]struct {
		args       args
		beforeTest func(a *args)
		wantError  error
		wantRc     int
	}{
		"countAllDocuments": {
			args: args{},
			beforeTest: func(a *args) {
				a.option = new(audit.ReadLogOption)
				start := time.Now().Add(-time.Minute * 5).UTC()
				a.option.StartTime = &start
				end := time.Now().Add(time.Minute * 5).UTC()
				a.option.EndTime = &end
				// log.Printf("time.Now(): %s", time.Now())
				// log.Printf("start: %s", a.option.StartTime)
				// log.Printf("end: %s", a.option.EndTime)
				for j := 0; j < 9; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
					err := db.Log(&a.events[j])
					require.NoError(t, err) //stops test execution if fail
				}
				exportCollection(*db, db.collection, fmt.Sprintf("%s/%s.json", db.path, db.collection))
			},
			wantError: nil,
			wantRc:    9,
		},
		"countOldDocuments": {
			args: args{},
			beforeTest: func(a *args) {
				a.option = new(audit.ReadLogOption)
				start := time.Now().UTC()
				a.option.StartTime = &start
				end := start.Add(time.Second * 15).UTC()
				a.option.EndTime = &end
				log.Printf("time.Now().UTC(): %s", time.Now().UTC())
				log.Printf("start: %s", a.option.StartTime)
				log.Printf("end: %s", a.option.EndTime)
				for j := 0; j < 9; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
					err := db.Log(&a.events[j])
					require.NoError(t, err) //stops test execution if fail
					time.Sleep(time.Second * 3)
				}
				exportCollection(*db, db.collection, fmt.Sprintf("%s/%s.json", db.path, db.collection))
			},
			wantError: nil,
			wantRc:    5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t, name, db)
			defer teardownTest(t, name, db)
			require.NoError(t, err)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}

			size, err := db.TotalCount(tt.args.option)
			if err != nil {
				t.Errorf("Can not read audit log; got %v", err)
			}
			if size != tt.wantRc {
				t.Errorf("Wrong number of registers; got %v, want %v", size, tt.wantRc)
			}
		})
	}
}

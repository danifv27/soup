package clover

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (func(t *testing.T), error) {
	log.Println("setup test")
	// Return a function to teardown the suite
	return func(t *testing.T) {
		log.Println("teardown test")
	}, nil
}

// You can use testing.T, if you want to test the code without benchmarking
func setupSuite(t *testing.T) (func(t *testing.T, db *CloverAuditer), *CloverAuditer) {
	log.Println("setup suite")

	dir, err := ioutil.TempDir("", "clover-test")
	require.NoError(t, err) //stops test execution if fail

	c, err := NewCloverAuditer(dir, "test")
	require.NoError(t, err) //stops test execution if fail

	// Return a function to teardown the test
	return func(t *testing.T, db *CloverAuditer) {
		log.Println("teardown suite")
		db.Close()
	}, c
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
		"insert one document": {
			args: args{},
			beforeTest: func(a *args) {
				a.events = make([]audit.Event, 0, 7)
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
		"insert multiple documents": {
			args: args{},
			beforeTest: func(a *args) {
				a.events = make([]audit.Event, 0, 7)
				for j := 0; j < 7; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/develop-%d", j),
					})
				}
			},
			wantError: nil,
			wantRc:    8, //Because we do not remove elements from previous test
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t)
			defer teardownTest(t)
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
			regs, err := db.Count()
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
		"read all documents": {
			args: args{},
			beforeTest: func(a *args) {
				a.events = make([]audit.Event, 0, 7)
				for j := 0; j < 9; j++ {
					a.events = append(a.events, audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/master-%d", j),
					})
					err := db.Log(&a.events[j])
					require.NoError(t, err) //stops test execution if fail
				}
			},
			wantError: nil,
			wantRc:    9,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t)
			defer teardownTest(t)
			require.NoError(t, err)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}
			events, err := db.ReadLog(nil)
			if err != nil {
				t.Errorf("Can not read audit log; got %v", err)
			}
			if len(events) != tt.wantRc {
				t.Errorf("Wrong number of registers; got %v, want %v", len(events), tt.wantRc)
			}
		})
	}
}

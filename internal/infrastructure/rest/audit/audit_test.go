package audit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/infrastructure/audit/clover"
	"github.com/danifv27/soup/internal/infrastructure/logger/logrus"
	"github.com/stretchr/testify/require"
)

type suite struct {
	db clover.CloverAuditer
	h  Handler
}

func setupSuite(t *testing.T) (func(t *testing.T, s suite), suite) {
	var s suite
	var err error

	dir, err := ioutil.TempDir("", "clover-test")
	require.NoError(t, err) //stops test execution if fail
	log.Printf("setting up suite (dir: %s)", dir)
	s.db, err = clover.NewCloverAuditer(fmt.Sprintf("audit:clover?path=%s&collection=test", dir))
	require.NoError(t, err) //stops test execution if fail
	s.h.LoggerService = logrus.NewLoggerService()
	s.h.Auditer = s.db

	return func(t *testing.T, s suite) {
		log.Println("teardown suite")
		s.db.Close()
	}, s
}

func setupTest(t *testing.T, name string, s suite) (func(t *testing.T, name string, s suite), error) {
	log.Printf("setting up test: %s", name)
	err := s.db.UseCollection(name)
	if err != nil {
		size, _ := s.db.GetNumberOfEvents(nil)
		log.Printf("collection %s created, size %d", name, size)
	}
	// Return a function to teardown the suite
	return func(t *testing.T, name string, s suite) {
		log.Printf("teardown test: %s", name)
		// Check if collection already exists
		collectionExists, _ := s.db.HasCollection(name)
		if collectionExists {
			s.db.DropCollection(name)
			log.Printf("collection %s dropped", name)
		}
	}, nil
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetEvents(t *testing.T) {
	teardownSuite, s := setupSuite(t)
	defer teardownSuite(t, s)

	type args struct {
		// option *audit.GetEventOption
		req *http.Request
	}

	tests := map[string]struct {
		args             args
		beforeTest       func(a *args)
		wantResponseCode int
		wantNumberRegs   int
	}{
		"checkJsonFormat": {
			args: args{},
			beforeTest: func(a *args) {
				// a.option = new(audit.GetEventOption)
				start := time.Now().Add(-time.Minute * 5).UTC()
				// a.option.StartTime = &start
				end := time.Now().Add(time.Minute * 5).UTC()
				// a.option.EndTime = &end
				req, err := http.NewRequest("GET", fmt.Sprintf("/audit?from=%s&to=%s", start.Format(time.RFC3339), end.Format(time.RFC3339)), nil)
				require.NoError(t, err) //stops test execution if fail
				a.req = req
				for j := 0; j < 2; j++ {
					evt := audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/develop-%d", j),
					}
					err := s.h.Auditer.Audit(&evt)
					require.NoError(t, err)
				}
			},
			wantResponseCode: http.StatusOK,
			wantNumberRegs:   2,
		},
		"checkLimitParam": {
			args: args{},
			beforeTest: func(a *args) {
				// a.option = new(audit.GetEventOption)
				start := time.Now().Add(-time.Minute * 5).UTC()
				// a.option.StartTime = &start
				end := time.Now().Add(time.Minute * 5).UTC()
				// a.option.EndTime = &end
				req, err := http.NewRequest("GET", fmt.Sprintf("/audit?from=%s&to=%s&limit=6", start.Format(time.RFC3339), end.Format(time.RFC3339)), nil)
				require.NoError(t, err) //stops test execution if fail
				a.req = req
				for j := 0; j < 14; j++ {
					evt := audit.Event{
						Action:  "repo:refs_changed",
						Actor:   "fraildan",
						Message: fmt.Sprintf("refs/heads/develop-%d", j),
					}
					err := s.h.Auditer.Audit(&evt)
					require.NoError(t, err)
				}
			},
			wantResponseCode: http.StatusOK,
			wantNumberRegs:   6,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			teardownTest, err := setupTest(t, name, s)
			require.NoError(t, err)
			defer teardownTest(t, name, s)
			if tt.beforeTest != nil {
				tt.beforeTest(&tt.args)
			}
			w := httptest.NewRecorder()
			s.h.GetEvents(w, tt.args.req)
			res := w.Result()
			defer res.Body.Close()
			checkResponseCode(t, http.StatusOK, res.StatusCode)
			got, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			// Convert the JSON response to a map
			var m []map[string]string
			err = json.Unmarshal(got, &m)
			require.NoError(t, err)
			if len(m) != tt.wantNumberRegs {
				t.Errorf("Wrong number of audit registers; got %v, want %v", len(m), tt.wantNumberRegs)
			}
		})
	}
}

package audit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
)

type Handler struct {
	apps application.Applications
}

//NewHandler Constructor
func NewHandler(app application.Applications) *Handler {

	return &Handler{
		apps: app,
	}
}

func isRFC3339V2(datetime string) (bool, time.Time, error) {
	var t time.Time
	var err error

	if t, err = time.Parse(time.RFC3339, datetime); err != nil {
		return false, time.Now().UTC(), err
	}

	loc := t.Location()

	return (loc == time.UTC), t, nil
}

//GetEvents Returns audit events
func (c Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	var option audit.GetEventOption
	var start, end time.Time
	var err error
	var isRFC bool
	var events []audit.Event

	from := r.URL.Query().Get("from")
	if from == "" {
		c.apps.LoggerService.Info("Missing from query argument")
		//https://stackoverflow.com/questions/3050518/what-http-status-response-code-should-i-use-if-the-request-is-missing-a-required
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	isRFC, start, err = isRFC3339V2(from)
	if err != nil {
		c.apps.LoggerService.WithFields(logger.Fields{
			"err":  err.Error(),
			"from": from,
		}).Info("Malformed query argument")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isRFC {
		c.apps.LoggerService.WithFields(logger.Fields{
			"err":  fmt.Sprintf("%s is not RFC3339 compliant", from),
			"from": from,
		}).Info("Wrong RFC3339 format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	to := r.URL.Query().Get("to")
	if to == "" {
		end = time.Now().UTC()
	} else {
		isRFC, end, err = isRFC3339V2(to)
		if err != nil {
			c.apps.LoggerService.WithFields(logger.Fields{
				"err": err.Error(),
				"to":  to,
			}).Info("Malformed query argument")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isRFC {
			c.apps.LoggerService.WithFields(logger.Fields{
				"err": fmt.Sprintf("%s is not RFC3339 compliant", to),
				"to":  to,
			}).Info("Malformed query argument")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		option.Limit, err = strconv.Atoi(limit)
		if err != nil {
			c.apps.LoggerService.WithFields(logger.Fields{
				"err":   err.Error(),
				"limit": limit,
			}).Info("Malformed query argument")
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	option.StartTime = &start
	option.EndTime = &end
	enc := json.NewEncoder(w)
	events, err = c.apps.Auditer.GetEvents(&option)
	if err != nil {
		c.apps.LoggerService.WithFields(logger.Fields{
			"err": err.Error(),
		}).Info("Can not retrieve audit events")
		w.WriteHeader(http.StatusInternalServerError)
	}
	c.apps.LoggerService.Debug("Audit events retrieved")
	enc.Encode(events)
}

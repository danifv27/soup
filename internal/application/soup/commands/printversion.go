package commands

import (
	"encoding/json"
	"fmt"

	"github.com/danifv27/soup/internal/domain/soup"
)

type PrintVersionRequest struct {
	Format string
}

type PrintVersionRequestHandler interface {
	Handle(command PrintVersionRequest) error
}

type printVersionRequestHandler struct {
	repo soup.Version
}

//NewUpdateCragRequestHandler Constructor
func NewPrintVersionRequestHandler(version soup.Version) PrintVersionRequestHandler {

	return printVersionRequestHandler{
		repo: version,
	}
}

//Handle Handles the update request
func (h printVersionRequestHandler) Handle(command PrintVersionRequest) error {
	var info *soup.VersionInfo
	var err error
	var out []byte

	if info, err = h.repo.GetVersionInfo(); err != nil {
		return fmt.Errorf("Handle: %w", err)
	}
	if command.Format == "json" {
		if out, err = json.MarshalIndent(info, "", "    "); err != nil {
			return fmt.Errorf("Handle: %w", err)
		}
		fmt.Println(string(out))
	} else {
		fmt.Println(info)
	}

	return nil
}

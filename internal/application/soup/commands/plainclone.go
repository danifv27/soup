package commands

import (
	"github.com/danifv27/soup/internal/domain/soup"
)

type PlainCloneRequest struct {
}

type PlainCloneRequestHandler interface {
	Handle(command PlainCloneRequest) error
}

type plainCloneRequestHandler struct {
	repo soup.Git
}

//NewUpdateCragRequestHandler Constructor
func NewPlainCloneRequestHandler(git soup.Git) PlainCloneRequestHandler {

	return plainCloneRequestHandler{repo: git}
}

//Handle Handles the update request
func (h plainCloneRequestHandler) Handle(command PlainCloneRequest) error {
	// crag, err := h.repo.PlainClone(command.ID)
	// if crag == nil {
	// 	return fmt.Errorf("the provided crag id does not exist")
	// }
	// if err != nil {
	// 	return err
	// }

	// crag.Name = command.Name
	// crag.Desc = command.Desc
	// crag.Country = command.Country

	// return h.repo.Update(*crag)
	return nil
}

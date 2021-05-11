package server

import (
	"errors"
	"sort"

	"github.com/ewagmig/rewards-collection/utils"
)

// Router defines an individual API route in the API server.
type Router struct {
	Path         string
	Method       string
	AuthType     utils.AuthType
	AllowedRoles []string
	Handler      utils.HandlerFunc
}

// Controller represents an interface to specify a group of routes.
type Controller interface {
	// Name returns the name to identity the controller uniquely.
	Name() string
	// Routes returns the list of routes in this controller.
	Routes() []*Router
}

var (
	errDuplicateCtrl = errors.New("Controller exists")
)

// ControllerList defines a group of controllers.
type ControllerList []Controller

func (c ControllerList) Len() int           { return len(c) }
func (c ControllerList) Less(i, j int) bool { return c[i].Name() < c[j].Name() }
func (c ControllerList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var controllerList ControllerList

// RegisterController registers a controller to the controller list. So if you want to make a
// controller accessable, you should register it first.
func RegisterController(ctrl Controller) error {
	for _, c := range controllerList {
		if c.Name() == ctrl.Name() {
			return errDuplicateCtrl
		}
	}

	controllerList = append(controllerList, ctrl)
	sort.Sort(controllerList)
	return nil
}

// Controllers returns all available controllers.
func Controllers() []Controller {
	return controllerList
}

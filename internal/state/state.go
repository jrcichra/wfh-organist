package state

import (
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/parser/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type State struct {
	db      *gorm.DB
	profile string
	config  *config.Config
}

type Group struct {
	gorm.Model
	Name string
}

type Stop struct {
	gorm.Model
	Name    string
	Code    string
	Pressed bool
	GroupID int
	Group   Group
}

type APIStop struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Pressed bool   `json:"pressed"`
}

type APIGroup []APIStop

type APIState []map[string]APIGroup

func (s *State) Open(profile string) {
	var err error
	s.db, err = gorm.Open(sqlite.Open(profile+"/state.db"), &gorm.Config{})
	common.Must(err)
	s.config = &config.Config{}
	s.config.Read(profile)
	s.reconcile()
}

// Reconcile - reads the stops in from the YAML file and reconciles the database state to match
func (s *State) reconcile() {

	s.db.AutoMigrate(&Stop{})
	s.db.AutoMigrate(&Group{})

	// TODO: clear out stops and stop groups that are not in the YAML file

	// Make sure there's a row for each stop & group
	for _, top := range s.config.Stops {
		for groupName, group := range top {
			g := Group{Name: groupName}
			s.db.FirstOrCreate(&g, g)
			for _, stop := range group {
				s.db.FirstOrCreate(&Stop{Name: stop.Name, Code: stop.Code, Group: g}, Stop{Name: stop.Name})
			}
		}
	}
}

func (s *State) GetStopsForAPI() *APIState {

	type APIStopWithGroup struct {
		APIStop
		Group string
	}

	var apiStops []APIStopWithGroup
	s.db.Model(&Stop{}).Select([]string{"stops.id", "stops.name", "stops.pressed", "groups.name as \"group\""}).Joins("JOIN groups ON stops.group_id = groups.id").Order("stops.created_at").Find(&apiStops)

	// get unique list of groups from apiStops
	var groups []string
	for _, apiStop := range apiStops {
		if !common.Contains(groups, apiStop.Group) {
			groups = append(groups, apiStop.Group)
		}
	}

	var apiState APIState
	for _, group := range groups {
		var apiGroup []APIStop
		for _, stop := range apiStops {
			if stop.Group == group {
				apiGroup = append(apiGroup, stop.APIStop)
			}
		}
		apiState = append(apiState, map[string]APIGroup{group: apiGroup})
	}

	return &apiState
}

func (s *State) GetStop(id int) (string, bool) {
	type StopCodePressed struct {
		Code    string
		Pressed bool
	}
	var stopCodePressed StopCodePressed
	s.db.Model(&Stop{}).Select([]string{"code", "pressed"}).Where("id = ?", id).First(&stopCodePressed)
	return stopCodePressed.Code, stopCodePressed.Pressed
}

func (s *State) ToggleStop(id int) {
	// check if the stop is pressed
	type StopPressed struct {
		Pressed bool
	}
	var stop StopPressed
	s.db.Model(&Stop{}).Select([]string{"pressed"}).Where("id = ?", id).First(&stop)

	// toggle it in the database
	if stop.Pressed {
		s.db.Model(&Stop{}).Where("id = ?", id).Update("pressed", false)
	} else {
		s.db.Model(&Stop{}).Where("id = ?", id).Update("pressed", true)
	}

}

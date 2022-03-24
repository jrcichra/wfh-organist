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
	GroupID int
	Group   Group
}

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

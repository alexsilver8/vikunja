// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// SavedTaskView is a user's named snapshot of a project view's filter, search and sort.
// It is private to its owner and only stores what the client needs to re-apply it.
type SavedTaskView struct {
	ID            int64  `xorm:"autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique, numeric id of this saved view."`
	Title         string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The name shown in the saved views menu."`
	ProjectViewID int64  `xorm:"bigint not null INDEX" json:"project_view_id" readOnly:"true" doc:"The project view this saved view belongs to; taken from the URL."`
	OwnerID       int64  `xorm:"bigint not null INDEX" json:"-"`
	// Only used to scope the request to the project in the URL.
	ProjectID int64 `xorm:"-" json:"-"`

	Filter  string   `xorm:"text null" json:"filter" doc:"A filter query in the syntax of the task list's filter parameter."`
	Search  string   `xorm:"varchar(250) null" json:"search" valid:"runelength(0|250)" maxLength:"250" doc:"A free-text search applied together with the filter."`
	SortBy  []string `xorm:"JSON null" json:"sort_by" doc:"Task fields to sort by, in order of precedence."`
	OrderBy []string `xorm:"JSON null" json:"order_by" doc:"asc or desc for each sort_by entry at the same position."`

	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this saved view was created. You cannot change this value."`
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this saved view was last updated. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*SavedTaskView) TableName() string {
	return "saved_task_views"
}

func getSavedTaskViewByID(s *xorm.Session, id int64) (*SavedTaskView, error) {
	v := &SavedTaskView{}
	exists, err := s.Where("id = ?", id).Get(v)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSavedTaskViewDoesNotExist{SavedTaskViewID: id}
	}
	return v, nil
}

// validate rejects queries the task list would reject on apply, so a broken view can't be saved.
func (v *SavedTaskView) validate() error {
	if _, err := getTaskFiltersFromFilterString(v.Filter, ""); err != nil {
		return err
	}
	for i, field := range v.SortBy {
		param := &sortParam{
			sortBy:        field,
			orderBy:       orderInvalid,
			projectViewID: v.ProjectViewID,
		}
		if i < len(v.OrderBy) {
			param.orderBy = getSortOrderFromString(v.OrderBy[i])
		}
		if err := param.validate(); err != nil {
			return err
		}
	}
	v.OrderBy = v.OrderBy[:min(len(v.OrderBy), len(v.SortBy))]
	return nil
}

// ReadAll returns the current user's saved views of a project view, sorted by title.
func (v *SavedTaskView) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	view, err := GetProjectViewByIDAndProject(s, v.ProjectViewID, v.ProjectID)
	if err != nil {
		return nil, 0, 0, err
	}
	can, _, err := view.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	where := builder.Eq{
		"owner_id":        a.GetID(),
		"project_view_id": view.ID,
	}
	views := []*SavedTaskView{}
	err = s.
		Where(where).
		OrderBy("title ASC, id ASC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&views)
	if err != nil {
		return nil, 0, 0, err
	}

	total, err := s.Where(where).Count(&SavedTaskView{})
	return views, len(views), total, err
}

// Create saves a new view for the current user.
func (v *SavedTaskView) Create(s *xorm.Session, a web.Auth) error {
	if err := v.validate(); err != nil {
		return err
	}
	v.ID = 0
	v.OwnerID = a.GetID()
	_, err := s.Insert(v)
	return err
}

// Update renames a saved view or replaces its filter, search and sort.
func (v *SavedTaskView) Update(s *xorm.Session, _ web.Auth) error {
	if err := v.validate(); err != nil {
		return err
	}
	_, err := s.
		Where("id = ?", v.ID).
		Cols("title", "filter", "search", "sort_by", "order_by").
		Update(v)
	if err != nil {
		return err
	}
	updated, err := getSavedTaskViewByID(s, v.ID)
	if err != nil {
		return err
	}
	*v = *updated
	return nil
}

// Delete removes a saved view.
func (v *SavedTaskView) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.Where("id = ?", v.ID).Delete(&SavedTaskView{})
	return err
}

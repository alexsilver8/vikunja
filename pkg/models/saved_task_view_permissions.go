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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CanCreate lets anyone who can see the project view save a view of it.
func (v *SavedTaskView) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	view, err := GetProjectViewByIDAndProject(s, v.ProjectViewID, v.ProjectID)
	if err != nil {
		return false, err
	}
	can, _, err := view.CanRead(s, a)
	return can, err
}

// CanUpdate only lets the owner change a saved view.
func (v *SavedTaskView) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	// Loading into v would overwrite the new values sent by the client.
	return (&SavedTaskView{
		ID:            v.ID,
		ProjectViewID: v.ProjectViewID,
		ProjectID:     v.ProjectID,
	}).canDoSavedTaskView(s, a)
}

// CanDelete only lets the owner delete a saved view.
func (v *SavedTaskView) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return v.canDoSavedTaskView(s, a)
}

// Owners keep access after losing access to the project, so they can still clean up.
func (v *SavedTaskView) canDoSavedTaskView(s *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	stored, err := getSavedTaskViewByID(s, v.ID)
	if err != nil {
		return false, err
	}
	// Answer like a missing view so ids of other views or projects aren't confirmed.
	if stored.ProjectViewID != v.ProjectViewID {
		return false, ErrSavedTaskViewDoesNotExist{SavedTaskViewID: v.ID}
	}
	if _, err := GetProjectViewByIDAndProject(s, stored.ProjectViewID, v.ProjectID); err != nil {
		return false, ErrSavedTaskViewDoesNotExist{SavedTaskViewID: v.ID}
	}
	if stored.OwnerID != a.GetID() {
		return false, nil
	}
	projectID := v.ProjectID
	*v = *stored
	v.ProjectID = projectID
	return true, nil
}

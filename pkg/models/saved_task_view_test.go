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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestSavedTaskView_RemovedWithItsProjectView(t *testing.T) {
	u1 := &user.User{ID: 1}

	t.Run("deleting the project view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&ProjectView{ID: 1, ProjectID: 1}).Delete(s, u1))
		require.NoError(t, s.Commit())

		for _, id := range []int64{
			1,
			2,
			3,
		} {
			db.AssertMissing(t, "saved_task_views", map[string]interface{}{"id": id})
		}
		db.AssertExists(t, "saved_task_views", map[string]interface{}{
			"id":              4,
			"project_view_id": 2,
		}, false)
	})

	t.Run("purging the project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 45 is in user 1's trash and has no views in the fixtures.
		view := &ProjectView{
			ProjectID: 45,
			Title:     "List",
			ViewKind:  ProjectViewKindList,
		}
		_, err := s.Insert(view)
		require.NoError(t, err)
		saved := &SavedTaskView{
			Title:         "Purged with the project",
			ProjectViewID: view.ID,
			OwnerID:       u1.ID,
		}
		_, err = s.Insert(saved)
		require.NoError(t, err)

		project := &Project{ID: 45}
		can, err := project.CanDeletePermanently(s, u1)
		require.NoError(t, err)
		require.True(t, can)
		require.NoError(t, project.DeletePermanently(s, u1))
		require.NoError(t, s.Commit())

		db.AssertMissing(t, "saved_task_views", map[string]interface{}{"id": saved.ID})
		db.AssertExists(t, "saved_task_views", map[string]interface{}{"id": 1}, false)
	})
}

func TestSavedTaskView_DeletedWithItsOwner(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	require.NoError(t, DeleteUser(s, &user.User{ID: 2}))
	require.NoError(t, s.Commit())

	db.AssertMissing(t, "saved_task_views", map[string]interface{}{"id": 3})
	db.AssertExists(t, "saved_task_views", map[string]interface{}{"id": 1}, false)
}

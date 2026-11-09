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

package webtests

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture topology (pkg/db/fixtures/saved_task_views.yml):
//   - #1 "Due soon", #2 "All open": testuser1's, on list view 1 of project 1.
//   - #3: user2's, on the same view 1 — never visible or writable for testuser1.
//   - #4: testuser1's, on view 2 of project 1 — not part of view 1's list.
//
// Project 2 (view 5) is user3's with no share to testuser1; project 9 (view 33)
// is shared to testuser1 read-only, which is enough to save views of it.
func TestHumaSavedTaskView(t *testing.T) {
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/views/1/saved-views",
		idParam:  "id",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// Each setupTestEnv() regenerates the JWT secret, so the harnesses share one Echo.
	otherView := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/views/2/saved-views",
		idParam:  "id",
		t:        t,
		e:        owned.e,
	}
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/views/5/saved-views",
		idParam:  "id",
		t:        t,
		e:        owned.e,
	}
	viewOfOtherProject := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/views/5/saved-views",
		idParam:  "id",
		t:        t,
		e:        owned.e,
	}
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/9/views/33/saved-views",
		idParam:  "id",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("only own views of this view, by title", func(t *testing.T) {
			rec, err := owned.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Equal(t, []int64{2, 1}, labelIDsFromReadAll(t, rec.Body.Bytes()), rec.Body.String())
			assert.Contains(t, rec.Body.String(), `"sort_by":["due_date","id"]`)
			assert.NotContains(t, rec.Body.String(), `owner_id`)
		})
		t.Run("forbidden project", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("view of another project", func(t *testing.T) {
			_, err := viewOfOtherProject.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{
				"title": "Urgent",
				"filter": "priority >= 4",
				"search": "release",
				"sort_by": ["priority", "id"],
				"order_by": ["desc", "desc"],
				"project_view_id": 5
			}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Urgent"`)
			assert.Contains(t, rec.Body.String(), `"project_view_id":1`)
			db.AssertExists(t, "saved_task_views", map[string]interface{}{
				"title":           "Urgent",
				"project_view_id": 1,
				"owner_id":        1,
				"filter":          "priority >= 4",
				"search":          "release",
			}, false)
		})
		t.Run("read access is enough", func(t *testing.T) {
			_, err := readShared.testCreateWithUser(nil, nil, `{"title": "Mine"}`)
			require.NoError(t, err)
			db.AssertExists(t, "saved_task_views", map[string]interface{}{
				"title":           "Mine",
				"project_view_id": 33,
				"owner_id":        1,
			}, false)
		})
		t.Run("position sort is allowed", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"title": "Manual", "sort_by": ["position"], "order_by": ["asc"]}`)
			require.NoError(t, err)
		})
		t.Run("forbidden project", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"title": "Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("view of another project", func(t *testing.T) {
			_, err := viewOfOtherProject.testCreateWithUser(nil, nil, `{"title": "Nope"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("empty title", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"title": ""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		for name, body := range map[string]string{
			"invalid filter":     `{"title": "Bad", "filter": "done = "}`,
			"unknown sort field": `{"title": "Bad", "sort_by": ["nope"], "order_by": ["asc"]}`,
			"invalid order":      `{"title": "Bad", "sort_by": ["id"], "order_by": ["sideways"]}`,
			"missing order":      `{"title": "Bad", "sort_by": ["id"]}`,
		} {
			t.Run(name, func(t *testing.T) {
				_, err := owned.testCreateWithUser(nil, nil, body)
				require.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("rename", func(t *testing.T) {
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"id": "1"}, `{
				"title": "Due this week",
				"filter": "done = false",
				"sort_by": ["due_date", "id"],
				"order_by": ["asc", "desc"]
			}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Due this week"`)
			assert.Contains(t, rec.Body.String(), `"filter":"done = false"`)
			db.AssertExists(t, "saved_task_views", map[string]interface{}{
				"id":       1,
				"title":    "Due this week",
				"owner_id": 1,
			}, false)
		})
		t.Run("someone else's view", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"id": "3"}, `{"title": "Taken over"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			db.AssertExists(t, "saved_task_views", map[string]interface{}{
				"id":    3,
				"title": "Someone else's view",
			}, false)
		})
		t.Run("view in the URL is not the saved view's", func(t *testing.T) {
			_, err := otherView.testUpdateWithUser(nil, map[string]string{"id": "1"}, `{"title": "Moved"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("nonexistent", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"id": "9999"}, `{"title": "Ghost"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("invalid sort", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"id": "1"}, `{"title": "Bad", "sort_by": ["nope"], "order_by": ["asc"]}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("someone else's view", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"id": "3"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			db.AssertExists(t, "saved_task_views", map[string]interface{}{"id": 3}, false)
		})
		t.Run("project in the URL is not the saved view's", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"id": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			db.AssertExists(t, "saved_task_views", map[string]interface{}{"id": 1}, false)
		})
		t.Run("normal", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"id": "2"})
			require.NoError(t, err)
			db.AssertMissing(t, "saved_task_views", map[string]interface{}{"id": 2})
		})
	})
}

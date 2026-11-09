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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type savedTaskViewListBody struct {
	Body Paginated[*models.SavedTaskView]
}

func RegisterSavedTaskViewRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "saved-task-views-list",
		Summary:     "List your saved views of a project view",
		Description: "Returns the saved views the authenticated user created for this project view, sorted by title. Saved views are private: other users' saved views are never returned. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views/{view}/saved-views",
		Tags:        tags,
	}, savedTaskViewsList)

	Register(api, huma.Operation{
		OperationID: "saved-task-views-create",
		Summary:     "Save a view of a project view",
		Description: "Saves a named filter, search and sort for this project view. Only the creator can see it. Requires read access to the project. The filter and sort are validated like the task list's parameters.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/views/{view}/saved-views",
		Tags:        tags,
	}, savedTaskViewsCreate)

	Register(api, huma.Operation{
		OperationID: "saved-task-views-update",
		Summary:     "Update a saved view",
		Description: "Replaces a saved view's title, filter, search and sort. Only its creator may update it.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/views/{view}/saved-views/{id}",
		Tags:        tags,
	}, savedTaskViewsUpdate)

	Register(api, huma.Operation{
		OperationID: "saved-task-views-delete",
		Summary:     "Delete a saved view",
		Description: "Deletes a saved view. Only its creator may delete it. Tasks are not affected.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/views/{view}/saved-views/{id}",
		Tags:        tags,
	}, savedTaskViewsDelete)
}

func init() { AddRouteRegistrar(RegisterSavedTaskViewRoutes) }

func savedTaskViewsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	ListParams
}) (*savedTaskViewListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	model := &models.SavedTaskView{
		ProjectID:     in.ProjectID,
		ProjectViewID: in.ViewID,
	}
	result, _, total, err := handler.DoReadAll(ctx, model, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.SavedTaskView)
	if !ok {
		return nil, fmt.Errorf("savedTaskViews.ReadAll returned unexpected type %T (expected []*models.SavedTaskView)", result)
	}
	return &savedTaskViewListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func savedTaskViewsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	Body      models.SavedTaskView
}) (*singleBody[models.SavedTaskView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	v := &in.Body
	v.ProjectID = in.ProjectID  // URL wins over body
	v.ProjectViewID = in.ViewID // URL wins over body
	if err := handler.DoCreate(ctx, v, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.SavedTaskView]{Body: v}, nil
}

func savedTaskViewsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	ID        int64 `path:"id"`
	Body      models.SavedTaskView
}) (*singleBody[models.SavedTaskView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	v := &in.Body
	v.ID = in.ID                // URL wins over body
	v.ProjectID = in.ProjectID  // URL wins over body
	v.ProjectViewID = in.ViewID // URL wins over body
	if err := handler.DoUpdate(ctx, v, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.SavedTaskView]{Body: v}, nil
}

func savedTaskViewsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	ID        int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	v := &models.SavedTaskView{
		ID:            in.ID,
		ProjectID:     in.ProjectID,
		ProjectViewID: in.ViewID,
	}
	if err := handler.DoDelete(ctx, v, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

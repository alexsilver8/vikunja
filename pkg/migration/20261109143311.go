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

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type savedTaskViews20261109143311 struct {
	ID            int64     `xorm:"autoincr not null unique pk"`
	Title         string    `xorm:"varchar(250) not null"`
	ProjectViewID int64     `xorm:"bigint not null INDEX"`
	OwnerID       int64     `xorm:"bigint not null INDEX"`
	Filter        string    `xorm:"text null"`
	Search        string    `xorm:"varchar(250) null"`
	SortBy        []string  `xorm:"JSON null"`
	OrderBy       []string  `xorm:"JSON null"`
	Created       time.Time `xorm:"created not null"`
	Updated       time.Time `xorm:"updated not null"`
}

func (savedTaskViews20261109143311) TableName() string {
	return "saved_task_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261109143311",
		Description: "Add saved task views",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(savedTaskViews20261109143311{}) //nolint:forbidigo // brand-new table, nothing to drop
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(savedTaskViews20261109143311{})
		},
	})
}

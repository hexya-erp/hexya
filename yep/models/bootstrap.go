/*   Copyright (C) 2008-2016 by Nicolas Piganeau and the TS2 team
 *   (See AUTHORS file)
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the
 *   Free Software Foundation, Inc.,
 *   59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
 */

package models

import (
	"fmt"
	"github.com/npiganeau/yep/yep/orm"
	"time"
)

type Option int

type BaseModel struct {
	ID         int64     `orm:"column(id)"`
	CreateDate time.Time `orm:"auto_now_add"`
	CreateUid  int64
	WriteDate  time.Time `orm:"auto_now"`
	WriteUid   int64
}

type BaseTransientModel struct {
	ID int64 `orm:"column(id)"`
}

const (
	TRANSIENT_MODEL Option = 1 << iota
)

func CreateModel(name string, options ...Option) {
	var opts Option
	for _, o := range options {
		opts |= o
	}
	var model interface{}
	if opts&TRANSIENT_MODEL > 0 {
		model = new(BaseTransientModel)
	} else {
		model = new(BaseModel)
	}
	orm.RegisterModelWithName(name, model)
	registerModelFields(name, model)
}

func ExtendModel(name string, models ...interface{}) {
	orm.RegisterModelExtension(name, models...)
	for _, model := range models {
		registerModelFields(name, model)
	}
}

/*
BootStrap freezes model, fields and method caches and syncs the database structure
with the declared data.
*/
func BootStrap() {
	err := orm.RunSyncdb("default", true, orm.Debug)
	if err != nil {
		panic(fmt.Errorf("Unable to sync database: %s", err))
	}
	methodsCache.Lock()
	defer methodsCache.Unlock()
	methodsCache.done = true

	fieldsCache.Lock()
	defer fieldsCache.Unlock()
	processDepends()

	fieldsCache.done = true
}

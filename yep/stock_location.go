/*   Copyright (C) 2008-2016 by Nicolas Piganeau
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

package yep

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
)

type BaseStockLocation struct {
	YepModel
	Name string
}

func (bsl BaseStockLocation) PrintName() {
	fmt.Printf("Struct: %T, Name: %s\n", bsl, bsl.Name)
}

type YepModel struct {
	gorm.Model
}

func (ym *YepModel) Super(t interface{}, fName string) {
	fmt.Printf("Type: %s, function: %v\n", reflect.ValueOf(t).Type(), fName)
	//	subStructValue := reflect.ValueOf(t).Elem().Field(0)
	//	for ssv := subStructValue; ssv.Kind() == reflect.Struct; ssv = ssv.Field(0) {
	//		fmt.Printf("ssv: %s, ssv.numMethod: %d\n", ssv.Type(), ssv.NumMethod())
	//		if ssv.MethodByName(fName).Kind() != reflect.Invalid {
	//			ssv.MethodByName(fName).Call([]reflect.Value{})
	//			return
	//		}
	//	}
	//	fmt.Println("Sortie sans trouver la solution")

	subStructValue := reflect.ValueOf(t).Field(0)
	subFunction := subStructValue.MethodByName(fName)
	subFunction.Call([]reflect.Value{})
}

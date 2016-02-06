/*   Copyright (C) 2016 by Nicolas Piganeau
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

package main

import (
	"fmt"
	//_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Hello world")
	//db, err := gorm.Open("postgres", "user=nicolas dbname=test_orm password=nicolas sslmode=disable")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//db.AutoMigrate(&sale.ResPartner{}, &base.ResUser{})
	//parent := sale.ResPartner{}
	//children := []*sale.ResPartner{}
	//db.Where("id=?", 2).Preload("Children", &children).Find(&parent)
	//fmt.Printf("Parent: %+v\n", parent)
	//fmt.Printf("Children: %v\n", children)
}

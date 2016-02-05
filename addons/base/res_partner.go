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

package base

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"image"
	"image/color"
	"time"
)

type BaseResPartnerTitle struct {
	gorm.Model
}

type ResPartnerTitle struct {
	BaseResPartnerTitle `yep:"include"`
}

type BaseResPartnerCategory struct {
	gorm.Model
}

type ResPartnerCategory struct {
	BaseResPartnerCategory `yep:"include"`
}

type BaseResPartner struct {
	gorm.Model
	Name             string
	Date             time.Time
	Title            ResPartnerTitle
	//Parent           ResPartner
	//Children         []ResPartner
	Ref              string
	Lang             string
	TZ               string
	TZOffset         string
	User             ResUser
	VAT              string
	Banks            []ResPartnerBank
	Website          string
	Comment          string
	Categories       []ResPartnerCategory
	CreditLimit      float64
	EAN13            string
	Active           bool
	Customer         bool
	Supplier         bool
	Employee         bool
	Function         string
	Type             string
	Street           string
	Street2          string
	Zip              string
	City             string
	State            ResCountryState
	Country          ResCountry
	Email            string
	Phone            string
	Fax              string
	Mobile           string
	Birthdate        time.Time
	IsCompany        bool
	UseParentAddress bool
	Image            image.Image
	Company          ResCompany
	Color            color.Color
	Users            []ResUser

	//'has_image': fields.function(_has_image, type="boolean"),
	//'company_id': fields.many2one('res.company', 'Company', select=1),
	//'color': fields.integer('Color Index'),
	//'user_ids': fields.one2many('res.users', 'partner_id', 'Users'),
	//'contact_address': fields.function(_address_display,  type='char', string='Complete Address'),
	//
	//# technical field used for managing commercial fields
	//'commercial_partner_id': fields.function(_commercial_partner_id, type='many2one', relation='res.partner', string='Commercial Entity', store=_commercial_partner_store_triggers)

}

type ResPartner struct {
	BaseResPartner `yep:"include"`
}

func (brp *BaseResPartner) DisplayName() string {
	return brp.Name
}

//func (brp *BaseResPartner) ParentName() string {
//	return brp.Parent.Name
//}

//func (brp *BaseResPartner) HasImage() bool {
//	return bool(brp.Image)
//}

func (brp *BaseResPartner) ContactAddress() string {
	return fmt.Sprintf("%s\n%s\n%s %s\n%s\n%s", brp.Street, brp.Street2, brp.Zip, brp.City, brp.State, brp.Country)
}

func (brp *BaseResPartner) CommercialPartner() *ResPartner {
	return ResPartner(brp)
}

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

package tests

import "testing"

func AssertTrue(t *testing.T, expr bool, msg string) {
	if !expr {
		t.Errorf("%v: expression is false", msg)
	}
}

func AssertEqual(t *testing.T, a interface{}, b interface{}, msg string) {
	if a != b {
		t.Errorf("%v: %v(%T) is not equal to %v(%T)", msg, a, a, b, b)
	}
}

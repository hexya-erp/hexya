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

// substituteRelatedFields returns a copy of the given fields slice with
// related fields substituted by their related field path. It also returns
// the list of substitutions to be given to resetRelatedFields.
func (rs *RecordSet) substituteRelatedFields(fields []string) ([]string, []KeySubstitution) {
	// We create a map to check if the substituted field already exists
	duplMap := make(map[string]bool, len(fields))
	for _, field := range fields {
		duplMap[field] = true
	}
	// Now we go for the substitution
	res := make([]string, len(fields))
	var substs []KeySubstitution
	for i, field := range fields {
		fi, ok := rs.mi.fields.get(field)
		if ok && fi.related() {
			res[i] = fi.relatedPath
			substs = append(substs, KeySubstitution{
				Orig: jsonizePath(rs.mi, fi.relatedPath),
				New:  fi.json,
				Keep: duplMap[fi.relatedPath],
			})
			continue
		}
		res[i] = field
	}
	return res, substs
}

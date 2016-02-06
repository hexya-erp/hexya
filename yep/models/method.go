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

package models

/*
MethodFunc is a function defined in a module that is to be added to a Model as a
layer to a method.

In YEP, methods are defined incrementally by adding MethodFunc to the MethodStack
of a Model. A MethodFunc should call Super() to access the MethodFunc underneath
in the stack.
 */
type MethodFunc func(RecordSet, ...interface{}) interface{}

/*
MethodStack is a stack of MethodFunc that represents a method in a Model. Models
always call the MethodFunc at the top of the stack when calling the method.
 */
type MethodStack []*MethodFunc

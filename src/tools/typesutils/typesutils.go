// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package typesutils

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/hexya-erp/hexya/src/tools/nbutils"
)

// RecordSet is an approximation of models.RecordSet so as not to import models.
type RecordSet interface {
	// ModelName returns the name of the model of this RecordSet
	ModelName() string
	// Ids returns the ids in this set of Records
	Ids() []int64
	// Len returns the number of records in this RecordSet
	Len() int
	// IsEmpty returns true if this RecordSet has no records
	IsEmpty() bool
	// IsNotEmpty returns true if this RecordSet has at least one record
	IsNotEmpty() bool
}

// IsZero returns true if the given value is the zero value of its type or nil
func IsZero(value interface{}) bool {
	if value == nil {
		return true
	}
	if rc, ok := value.(RecordSet); ok {
		return rc.IsEmpty()
	}
	val := reflect.ValueOf(value)
	return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
)

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

// AreEqual returns true if both given values are equal.
func AreEqual(arg1, arg2 interface{}) (bool, error) {
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	v2 := reflect.ValueOf(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}
	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
		default:
			return false, errBadComparison
		}
	} else {
		switch k1 {
		case boolKind:
			truth = v1.Bool() == v2.Bool()
		case complexKind:
			truth = v1.Complex() == v2.Complex()
		case floatKind:
			truth = v1.Float() == v2.Float()
		case intKind:
			truth = v1.Int() == v2.Int()
		case stringKind:
			truth = v1.String() == v2.String()
		case uintKind:
			truth = v1.Uint() == v2.Uint()
		default:
			panic("invalid kind")
		}
	}
	if truth {
		return true, nil
	}
	return false, nil
}

// IsLessThan returns true if arg1 is less than arg2
// It panics if the kind of arg1 or arg2 is not a basic kind.
func IsLessThan(arg1, arg2 interface{}) (bool, error) {
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	v2 := reflect.ValueOf(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}
	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() < 0 || uint64(v1.Int()) < v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() < uint64(v2.Int())
		default:
			return false, errBadComparison
		}
	} else {
		switch k1 {
		case boolKind, complexKind:
			return false, errBadComparisonType
		case floatKind:
			truth = v1.Float() < v2.Float()
		case intKind:
			truth = v1.Int() < v2.Int()
		case stringKind:
			truth = v1.String() < v2.String()
		case uintKind:
			truth = v1.Uint() < v2.Uint()
		default:
			panic("invalid kind")
		}
	}
	return truth, nil
}

// Convert the given value to the given Type. Set isRS to true if the value represents a RecordSet
//
// If the target type implements sql.Scanner, then the Scan method is used for the conversion.
func Convert(value interface{}, target interface{}, isRS bool) error {
	targetType := reflect.TypeOf(target).Elem()
	if targetType == reflect.TypeOf(value) {
		// If we already have the good type, don't do anything
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(value))
		return nil
	}
	switch {
	case value == nil:
		// value is nil, we keep target unchanged
		return nil
	case reflect.PtrTo(targetType).Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()):
		// the type implements sql.Scanner, so we call Scan
		valPtr := reflect.ValueOf(target)
		scanFunc := valPtr.MethodByName("Scan")
		inArgs := []reflect.Value{reflect.ValueOf(value)}
		res := scanFunc.Call(inArgs)
		if res[0].Interface() != nil {
			return fmt.Errorf("unable to scan into target Type: %v", res[0].Interface())
		}
	default:
		var err error
		switch {
		case isRS:
			err = getRelationFieldValue(value, target)
		default:
			err = getSimpleTypeValue(value, target)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// getSimpleTypeValue returns value as a reflect.Value with type of targetType
// It returns an error if the value cannot be converted to the target type
func getSimpleTypeValue(value interface{}, target interface{}) error {
	targetType := reflect.TypeOf(target).Elem()
	val := reflect.ValueOf(value)
	var res interface{}
	_, tIsBool := target.(*bool)
	_, tIsFloat32 := target.(*float32)
	_, tIsFloat64 := target.(*float64)
	typ := val.Type()
	switch {
	case tIsBool:
		res = !reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
	case typ.ConvertibleTo(targetType):
		res = val.Convert(targetType).Interface()
	case typ == reflect.TypeOf([]byte{}) && tIsFloat32:
		// backend may return floats as []byte when stored as numeric
		fval, err := strconv.ParseFloat(string(value.([]byte)), 32)
		if err != nil {
			return err
		}
		res = float32(fval)
	case typ == reflect.TypeOf([]byte{}) && tIsFloat64:
		// backend may return floats as []byte when stored as numeric
		fval, err := strconv.ParseFloat(string(value.([]byte)), 64)
		if err != nil {
			return err
		}
		res = fval
	default:
		return fmt.Errorf("impossible conversion of %v (%T) to %s", value, value, targetType)
	}

	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(res))
	return nil
}

// getRelationFieldValue returns value as a reflect.Value with type of targetType
// It returns an error if the value is not consistent with a relation field value
// (i.e. is not of type RecordSet or int64 or []int64)
func getRelationFieldValue(value interface{}, target interface{}) error {
	var (
		ids []int64
		res interface{}
	)
	// Step 1: set ids
	switch tValue := value.(type) {
	case RecordSet:
		ids = tValue.Ids()
	case []interface{}:
		if len(tValue) == 0 {
			ids = []int64{}
			break
		}
		return errors.New("non empty []interface{} given")
	case []int64:
		ids = tValue
	case *interface{}:
		ids = []int64{}
	default:
		nbValue, nbErr := nbutils.CastToInteger(tValue)
		if nbErr != nil {
			return fmt.Errorf("expected number value, got %v: %s", value, nbErr)
		}
		ids = []int64{nbValue}
	}
	// Step 2 convert to target type
	switch target.(type) {
	case *int64:
		if len(ids) > 0 {
			res = ids[0]
		} else {
			res = int64(0)
		}
	case *[]int64:
		res = ids
	default:
		return errors.New("non consistent type")
	}
	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(res))
	return nil
}

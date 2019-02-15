// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package generate

import (
	"fmt"
)

// specificMethodsHandlers are functions that populate the given modelData
// for specific methods.
var specificMethodsHandlers = map[string]func(modelData *modelData, depsMap *map[string]bool){
	"Search":           searchMethodHandler,
	"SearchByName":     searchByNameMethodHandler,
	"Create":           createMethodHandler,
	"Write":            writeMethodHandler,
	"Copy":             copyMethodHandler,
	"CartesianProduct": cartesianProductMethodHandler,
	"Sorted":           sortedMethodHandler,
	"Filtered":         filteredMethodHandler,
	"Aggregates":       aggregatesMethodHandler,
	"First":            firstMethodHandler,
	"All":              allMethodHandler,
}

// searchMethodHandler returns the specific methodData for the Search method.
func searchMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Search"
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("%s.%sCondition", PoolQueryPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("%s.%sCondition", PoolQueryPackage, modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name:           name,
		Doc:            fmt.Sprintf("// Search returns a new %sSet filtering on the current one with the additional given Condition", modelData.Name),
		ToDeclare:      false,
		Params:         "condition",
		ParamsWithType: fmt.Sprintf("condition %s.%sCondition", PoolQueryPackage, modelData.Name),
		ReturnAsserts:  fmt.Sprintf("resTyped := res.(models.RecordSet).Collection().Wrap(\"%s\").(%s)", modelData.Name, returnString),
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// createMethodHandler returns the specific methodData for the Create method.
func createMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Create"
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("%s.%sData", PoolInterfacesPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("%sData", modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// Create inserts a %s record in the database from the given data.
// Returns the created %sSet.`,
			modelData.Name, modelData.Name),
		ToDeclare:      false,
		Params:         "data",
		ParamsWithType: fmt.Sprintf("data %s.%sData", PoolInterfacesPackage, modelData.Name),
		ReturnAsserts:  fmt.Sprintf("resTyped := res.(models.RecordSet).Collection().Wrap(\"%s\").(%s)", modelData.Name, returnString),
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// writeMethodHandler returns the specific methodData for the Write method.
func writeMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Write"
	returnString := "bool"
	iReturnString := "bool"
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		IParamsTypes:  fmt.Sprintf("%sData", modelData.Name),
		ParamsTypes:   fmt.Sprintf("%s.%sData", PoolInterfacesPackage, modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// Write is the base implementation of the 'Write' method which updates
// %s records in the database with the given data.`, modelData.Name),
		ToDeclare:      false,
		Params:         "data",
		ParamsWithType: fmt.Sprintf("data %s.%sData", PoolInterfacesPackage, modelData.Name),
		ReturnAsserts:  "resTyped, _ := res.(bool)",
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// copyMethodHandler returns the specific methodData for the Copy method.
func copyMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Copy"
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("%s.%sData", PoolInterfacesPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("%sData", modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name:           name,
		Doc:            fmt.Sprintf(`// Copy duplicates the given %s record, overridding values with overrides.`, modelData.Name),
		ToDeclare:      false,
		Params:         "overrides",
		ParamsWithType: fmt.Sprintf("overrides %s.%sData", PoolInterfacesPackage, modelData.Name),
		ReturnAsserts:  fmt.Sprintf("resTyped := res.(models.RecordSet).Collection().Wrap(\"%s\").(%s)", modelData.Name, returnString),
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// searchByNameMethodHandler returns the specific methodData for the Search method.
func searchByNameMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "SearchByName"
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	(*depsMap)["github.com/hexya-erp/hexya/src/models/operator"] = true
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("string, operator.Operator, %s.%sCondition, int", PoolQueryPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("string, operator.Operator, %s.%sCondition, int", PoolQueryPackage, modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
	modelData.Methods = append(modelData.Methods, methodData{
		Name: name,
		Doc: fmt.Sprintf(`// SearchByName searches for %s records that have a display name matching the given
// "name" pattern when compared with the given "op" operator, while also
// matching the optional search condition ("additionalCond").
//
// This is used for example to provide suggestions based on a partial
// value for a relational field. Sometimes be seen as the inverse
// function of NameGet but it is not guaranteed to be.`, modelData.Name),
		ToDeclare:      false,
		Params:         "name, op, additionalCond, limit",
		ParamsWithType: fmt.Sprintf("name string, op operator.Operator, additionalCond %s.%sCondition, limit int", PoolQueryPackage, modelData.Name),
		ReturnAsserts:  fmt.Sprintf("resTyped := res.(models.RecordSet).Collection().Wrap(\"%s\").(%s)", modelData.Name, returnString),
		Returns:        "resTyped",
		ReturnString:   returnString,
		Call:           "Call",
	})
}

// firstMethodHandler returns the specific methodData for the First method.
func firstMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "First"
	returnString := fmt.Sprintf("%s.%sData", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("%sData", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
}

// allMethodHandler returns the specific methodData for the First method.
func allMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "All"
	returnString := fmt.Sprintf("[]%s.%sData", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("[]%sData", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
}

// cartesianProductMethodHandler returns the specific methodData for the CartesianProduct method.
func cartesianProductMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "CartesianProduct"
	returnString := fmt.Sprintf("[]%s.%sSet", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("[]%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("...%s.%sSet", PoolInterfacesPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("...%sSet", modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
}

// sortedMethodHandler returns the specific methodData for the Sorted method.
func sortedMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Sorted"
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("func(%s.%sSet, %sSet) bool", PoolInterfacesPackage, modelData.Name, modelData.Name),
		IParamsTypes:  fmt.Sprintf("func(%sSet, %sSet) bool", modelData.Name, modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
}

// filteredMethodHandler returns the specific methodData for the Sorted method.
func filteredMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	name := "Filtered"
	returnString := fmt.Sprintf("%s.%sSet", PoolInterfacesPackage, modelData.Name)
	iReturnString := fmt.Sprintf("%sSet", modelData.Name)
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          name,
		ParamsTypes:   fmt.Sprintf("func(%s.%sSet) bool", PoolInterfacesPackage, modelData.Name),
		IParamsTypes:  fmt.Sprintf("func(%sSet) bool", modelData.Name),
		ReturnString:  returnString,
		IReturnString: iReturnString,
	})
}

// aggregatesMethodHandler returns the specific methodData for the Aggregates method.
func aggregatesMethodHandler(modelData *modelData, depsMap *map[string]bool) {
	modelData.AllMethods = append(modelData.AllMethods, methodData{
		Name:          "Aggregates",
		ParamsTypes:   "...models.FieldNamer",
		IParamsTypes:  "...models.FieldNamer",
		ReturnString:  fmt.Sprintf("[]%s.%sGroupAggregateRow", PoolInterfacesPackage, modelData.Name),
		IReturnString: fmt.Sprintf("[]%sGroupAggregateRow", modelData.Name),
	})
}

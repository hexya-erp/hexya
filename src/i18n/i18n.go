// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"sort"
	"strings"

	"github.com/hexya-erp/hexya/src/models/types"
	"github.com/hexya-erp/hexya/src/tools/po"
)

const fieldSep string = "."

// Registry holds all the translation of the application
var Registry *TranslationsCollection

// A TranslationsCollection holds all the translations of the application
type TranslationsCollection struct {
	fieldDescription map[fieldRef]string
	fieldHelp        map[fieldRef]string
	fieldSelection   map[selectionRef]string
	resource         map[resourceRef]string
	code             map[codeRef]string
	custom           map[customRef]string
}

// TranslateFieldDescription returns the translation for the given model field
// name in the given lang. If no translation is found or if the translation
// is the empty string defaultValue is returned.
func (tc *TranslationsCollection) TranslateFieldDescription(lang, model, field, defaultValue string) string {
	key := fieldRef{lang: lang, model: model, field: field}
	val, ok := tc.fieldDescription[key]
	if !ok || val == "" {
		return defaultValue
	}
	return val
}

// TranslateFieldHelp returns the translation for the given model field
// help in the given lang. If no translation is found or if the translation
// is the empty string defaultValue is returned.
func (tc *TranslationsCollection) TranslateFieldHelp(lang, model, field, defaultValue string) string {
	key := fieldRef{lang: lang, model: model, field: field}
	val, ok := tc.fieldHelp[key]
	if !ok || val == "" {
		return defaultValue
	}
	return val
}

// TranslateFieldSelection returns the translated version of the given selection in the given lang.
// When no translation is found for an item, the original string is used.
func (tc *TranslationsCollection) TranslateFieldSelection(lang, model, field string, selection types.Selection) types.Selection {
	res := make(types.Selection)
	for selKey, selItem := range selection {
		key := selectionRef{lang: lang, model: model, field: field, source: selItem}
		val, ok := tc.fieldSelection[key]
		if !ok || val == "" {
			res[selKey] = selItem
			continue
		}
		res[selKey] = val
	}
	return res
}

// TranslateResourceItem returns the translation for the given src of the given resource
// in the given lang. If no translation is found or if the translation is the
// empty string src is returned.
func (tc *TranslationsCollection) TranslateResourceItem(lang, resourceID, src string) string {
	key := resourceRef{lang: lang, id: resourceID, source: src}
	val, ok := tc.resource[key]
	if !ok || val == "" {
		return src
	}
	return val
}

// TranslateCode returns the translation for the given src in the given lang, in the
// given context. If no translation is found or if the translation is the empty
// string src is returned.
func (tc *TranslationsCollection) TranslateCode(lang, context, src string) string {
	key := codeRef{lang: lang, context: context, source: src}
	val, ok := tc.code[key]
	if !ok || val == "" {
		return src
	}
	return val
}

// TranslateCustom returns the translation for the given src of the given custom po string
// in the given lang. If no translation is found or if the translation is the
// empty string src is returned.
func (tc *TranslationsCollection) TranslateCustom(lang, id, moduleName string) string {
	key := customRef{lang: lang, id: id, module: moduleName}
	val, ok := tc.custom[key]
	if !ok || val == "" {
		return id
	}
	return val
}

// LoadPOFile load the file with the given filename into the TranslationsCollection.
// This function can be called several times to iteratively load translations.
// It panics in case of errors in the PO file.
func (tc *TranslationsCollection) LoadPOFile(fileName string) {
	poFile, err := po.Load(fileName)
	if err != nil {
		log.Panic("Error while parsing PO file", "file", fileName, "error", err)
	}
	lang := poFile.MimeHeader.Language
	if lang == "" {
		log.Panic("Language should be specified in PO file header", "file", fileName)
	}
	for _, msg := range poFile.Messages {
		for _, line := range strings.Split(msg.ExtractedComment, "\n") {
			tokens := strings.Split(line, ":")
			if len(tokens) != 2 {
				log.Warn("Invalid format for PO comment. Should be '#. key:value'", "file", fileName, "line", msg.StartLine, "comment", line)
				continue
			}
			switch tokens[0] {
			case "field":
				// #. field:Model.Field
				meta := strings.Replace(tokens[1], " ", "", -1)
				r := strings.Split(meta, fieldSep)
				if len(r) != 2 {
					log.Panic("Invalid format for PO comment. Field reference should be 'Model.Field'", "file", fileName, "line", msg.StartLine, "comment", line)
				}
				tc.fieldDescription[fieldRef{lang: lang, model: r[0], field: r[1]}] = msg.MsgStr
			case "help":
				// #. help:Model.Field
				meta := strings.Replace(tokens[1], " ", "", -1)
				r := strings.Split(meta, fieldSep)
				if len(r) != 2 {
					log.Panic("Invalid format for PO comment. Field reference should be 'Model.Field'", "file", fileName, "line", msg.StartLine, "comment", line)
				}
				tc.fieldHelp[fieldRef{lang: lang, model: r[0], field: r[1]}] = msg.MsgStr
			case "selection":
				// #. selection:Model.Field
				meta := strings.Replace(tokens[1], " ", "", -1)
				r := strings.Split(meta, fieldSep)
				if len(r) != 2 {
					log.Panic("Invalid format for PO comment. Field reference should be 'Model.Field'", "file", fileName, "line", msg.StartLine, "comment", line)
				}
				tc.fieldSelection[selectionRef{lang: lang, model: r[0], field: r[1], source: msg.MsgId}] = msg.MsgStr
			case "resource":
				// #. resource:my_view_id
				viewID := strings.Replace(tokens[1], " ", "", -1)
				tc.resource[resourceRef{lang: lang, id: viewID, source: msg.MsgId}] = msg.MsgStr
			case "code":
				// #. code:
				// Translating code. Context may be given as msgctxt
				tc.code[codeRef{lang: lang, context: msg.MsgContext, source: msg.MsgId}] = msg.MsgStr
			case "custom":
				// #. custom: moduleName
				moduleName := strings.Replace(tokens[1], " ", "", -1)
				tc.custom[customRef{lang: lang, id: msg.MsgId, module: moduleName}] = msg.MsgStr
			}
		}
	}
}

// TranslateFieldDescription returns the translation for the given model field
// name in the given lang, using the default translation Registry. If no
// translation is found or if the translation is the empty string defaultValue
// is returned.
func TranslateFieldDescription(lang, model, field, defaultValue string) string {
	return Registry.TranslateFieldDescription(lang, model, field, defaultValue)
}

// TranslateFieldHelp returns the translation for the given model field
// help in the given lang, using the default translation Registry. If no
// translation is found or if the translation is the empty string defaultValue
// is returned.
func TranslateFieldHelp(lang, model, field, defaultValue string) string {
	return Registry.TranslateFieldHelp(lang, model, field, defaultValue)
}

// TranslateFieldSelection returns the translated version of the given selection
// in the given lang, using the default translation Registry. When no
// translation is found for an item, the original string is used.
func TranslateFieldSelection(lang, model, field string, selection types.Selection) types.Selection {
	return Registry.TranslateFieldSelection(lang, model, field, selection)
}

// TranslateResourceItem returns the translation for the given src of the given resource
// in the given lang using the default translation Registry. If no translation is found or if the translation is the
// empty string src is returned.
func TranslateResourceItem(lang, resourceID, src string) string {
	return Registry.TranslateResourceItem(lang, resourceID, src)
}

// TranslateCode returns the translation for the given src in the given lang, in the
// given context. If no translation is found or if the translation is the empty
// string src is returned.
func TranslateCode(lang, context, src string) string {
	return Registry.TranslateCode(lang, context, src)
}

// TranslateCustom returns the custom translation for the given id
func TranslateCustom(lang, id, moduleName string) string {
	return Registry.TranslateCustom(lang, id, moduleName)
}

// A fieldRef references a field in the translation maps
type fieldRef struct {
	lang  string
	model string
	field string
}

// A selectionRef references a selection item translation
type selectionRef struct {
	lang   string
	model  string
	field  string
	source string
}

// A resourceRef references a text translation in a resource
type resourceRef struct {
	lang   string
	id     string
	source string
}

// A codeRef references a translated text in code
type codeRef struct {
	lang    string
	context string
	source  string
}

// A customRef references a custom translated text
type customRef struct {
	lang   string
	id     string
	module string
}

// A Translation holds all the translations for a given language
type Translation struct {
	language string
}

// NewTranslationsCollection returns a pointer to a new TranslationsCollection ready for use
func NewTranslationsCollection() *TranslationsCollection {
	return &TranslationsCollection{
		fieldDescription: make(map[fieldRef]string),
		fieldHelp:        make(map[fieldRef]string),
		fieldSelection:   make(map[selectionRef]string),
		resource:         make(map[resourceRef]string),
		code:             make(map[codeRef]string),
		custom:           make(map[customRef]string),
	}
}

// LoadPOFile load the file with the given filename into the Registry.
// This function is meant to be called several times to load all translations
// across all languages and modules.
// It panics in case of errors in the PO file.
func LoadPOFile(fileName string) {
	Registry.LoadPOFile(fileName)
}

// GetAllCustomTranslations returns all custom translations by lang and by modules
func GetAllCustomTranslations() map[string]map[string]map[string]string {
	res := make(map[string]map[string]map[string]string)
	for key, val := range Registry.custom {
		if res[key.lang] == nil {
			res[key.lang] = make(map[string]map[string]string)
		}
		if res[key.lang][key.module] == nil {
			res[key.lang][key.module] = make(map[string]string)
		}
		res[key.lang][key.module][key.id] = val
	}
	return res
}

// allLanguageList is a cache for all locales ISO codes slice
var allLanguageList []string

func updateAllLanguageList() {
	allLanguageList = make([]string, len(locales))
	var i int
	for loc := range locales {
		allLanguageList[i] = loc
		i++
	}
	sort.Strings(allLanguageList)
}

// GetAllLanguageList returns a slice containing all known language codes
func GetAllLanguageList() []string {
	if allLanguageList == nil {
		updateAllLanguageList()
	}
	return allLanguageList
}

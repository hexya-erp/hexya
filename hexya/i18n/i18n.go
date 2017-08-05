// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"strings"

	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/po"
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
	key := resourceRef{lang: lang, viewID: resourceID, source: src}
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
				tc.resource[resourceRef{lang: lang, viewID: viewID, source: msg.MsgId}] = msg.MsgStr
			default:
				// Translating code. Context may be given as msgctxt
				tc.code[codeRef{lang: lang, context: msg.MsgContext, source: msg.MsgId}] = msg.MsgStr
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
	viewID string
	source string
}

// A codeRef references a translated text in code
type codeRef struct {
	lang    string
	context string
	source  string
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
	}
}

// LoadPOFile load the file with the given filename into the Registry.
// This function is meant to be called several times to load all translations
// across all languages and modules.
// It panics in case of errors in the PO file.
func LoadPOFile(fileName string) {
	Registry.LoadPOFile(fileName)
}

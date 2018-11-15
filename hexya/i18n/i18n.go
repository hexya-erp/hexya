// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package i18n

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/po"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
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

// loadLangParametersMap returns a map containing all required language informations.
// Those informations are read from data file in <HexPath>/hexya/i18n/data/langParameters.csv
func loadLangParametersMap() map[string]LangParameters {
	out := make(map[string]LangParameters)
	path := filepath.Join(generate.HexyaDir, "hexya", "i18n", "data", "langParameters.csv")
	r, err := os.Open(path)
	if err != nil {
		log.Panic("langParameters.csv Not found", "path", path)
	}
	reader := csv.NewReader(r)
	content, err := reader.ReadAll()
	if err != nil {
		log.Panic("Error while parsing csv file", "path", path, "err", err)
	}
	for _, line := range content[1:] {
		direction := LangDirectionLTR
		if line[4] == "Right-to-Left" {
			direction = LangDirectionRTL
		}
		out[line[3]] = LangParameters{
			Name:         line[1],
			Direction:    direction,
			DateFormat:   line[8],
			TimeFormat:   line[9],
			ThousandsSep: line[7],
			DecimalPoint: line[6],
			Grouping:     line[5],
		}
	}
	return out
}

// LangParametersMap contains all LangParameters loaded from data file
var langParametersMap map[string]LangParameters

// GetLangParameters returns a LangParameters struct describing a language's rules
// at first call, the data file containing all languages parameters is read
// if the language is not loaded, it returns a LangParameters similar to English (en_US)
func GetLangParameters(lang string) LangParameters {
	if langParametersMap == nil {
		langParametersMap = loadLangParametersMap()
	}
	out, ok := langParametersMap[lang]
	if !ok {
		return LangParameters{
			Name:         fmt.Sprintf("UNKNOWN_LOCALE (%s)", lang),
			Direction:    LangDirectionLTR,
			DateFormat:   `%m/%d/%Y`,
			TimeFormat:   `%H:%M:%S`,
			ThousandsSep: `,`,
			DecimalPoint: `.`,
			Grouping:     `[3,0]`,
		}
	}
	return out
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

var allLanguageList []string

// getLanguageListInFolder appends to a slice all language codes read in the given folder
func getLanguageListInFolder(out []string, path string) []string {
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() && p != path {
			return filepath.SkipDir
		}
		if strings.HasSuffix(p, ".po") {
			value := strings.TrimSuffix(filepath.Base(p), ".po")
			if !strutils.IsInStringSlice(value, out) {
				out = append(out, value)
			}
		}
		return nil
	})
	return out
}

// GetAllLanguageList returns a slice containing all known language codes
func GetAllLanguageList() []string {
	if allLanguageList == nil {
		out := []string{`af`, `am`, `ar`, `bg`, `bs`, `ca`, `cs`, `da`, `de`, `el`, `en_AU`, `en_GB`, `es`, `es_AR`,
			`es_BO`, `es_CL`, `es_CO`, `es_CR`, `es_DO`, `es_EC`, `es_PA`, `es_PE`, `es_PY`, `es_VE`, `et`, `eu`,
			`fa`, `fi`, `fo`, `fr`, `fr_BE`, `fr_CA`, `gl`, `gu`, `he`, `hr`, `hu`, `hy`, `id`, `is`, `it`, `ja`, `ka`,
			`kab`, `km`, `ko`, `lo`, `lt`, `lv`, `mk`, `mn`, `nb`, `ne`, `nl`, `nl_BE`, `pl`, `pt`, `pt_BR`, `ro`, `ru`,
			`sk`, `sl`, `sq`, `sr`, `sr@latin`, `sv`, `ta`, `th`, `tr`, `uk`, `vi`, `zh_CN`, `zh_TW`}
		path := filepath.Join(generate.HexyaDir, "server", "i18n", "*")
		symlinks, err := filepath.Glob(path)
		if err != nil {
			log.Error("Could not find any glob match", "path", path+"*", "error", err)
			symlinks = nil
		}
		for _, link := range symlinks {
			fi, _ := os.Lstat(link)
			if fi.Mode()&os.ModeSymlink == 0 {
				continue
			}
			path, err = os.Readlink(link)
			if err != nil {
				log.Warn("Could not read symlink", `link`, link, `error`, err)
				continue
			}
			out = getLanguageListInFolder(out, path)
		}
		sort.Slice(out, func(i, j int) bool {
			cmp := strings.Compare(out[i], out[j])
			return cmp < 0
		})
		allLanguageList = out
	}
	return allLanguageList
}

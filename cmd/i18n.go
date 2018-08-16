// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package cmd

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/po"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"github.com/hexya-erp/hexya/hexya/views"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/loader"
)

var i18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "Internationalization utilities",
	Long:  `Internationalization utilities for Hexya`,
}

var i18nUpdate = &cobra.Command{
	Use:   "update [dir]",
	Short: "Create or update PO files",
	Long: `Create or update PO files of the module specified by 'dir'.
PO files will be generated for each loaded language (--language flag)
in the i18n directory of the module.`,
	Run: func(cmd *cobra.Command, args []string) {
		moduleDir := "."
		if len(args) > 0 {
			moduleDir = args[0]
		}
		langs, err := cmd.Flags().GetStringSlice("languages")
		if err != nil {
			log.Panic("Unable to read languages from the command line")
		}
		updatePOFiles(moduleDir, langs)
	},
}

// A messageRef identifies unique messages
type messageRef struct {
	msgId   string
	msgCtxt string
}

// updatePOFiles creates or updates PO files of the module in the given
// dir with the data in the Translation registry.
func updatePOFiles(moduleDir string, langs []string) {
	i18nDir := filepath.Join(moduleDir, "i18n")
	server.LoadModuleTranslations(i18nDir, langs)
	conf := loader.Config{}
	conf.Import(moduleDir)
	program, err := conf.Load()
	if err != nil {
		log.Panic("Unable to build program", "error", err)
	}
	packs := program.InitialPackages()
	if len(packs) != 1 {
		log.Panic("Something has gone wrong, we have more than one package", "packs", packs)
	}
	modInfos := []*generate.ModuleInfo{{PackageInfo: *packs[0], ModType: generate.Base}}
	modelsASTData := generate.GetModelsASTDataForModules(modInfos, false)

	for _, lang := range langs {
		messages := make(map[messageRef]po.Message)
		for model, modelASTData := range modelsASTData {
			for field, fieldASTData := range modelASTData.Fields {
				messages = addDescriptionToMessages(lang, model, field, fieldASTData, messages)
				messages = addHelpToMessages(lang, model, field, fieldASTData, messages)
				messages = addSelectionToMessages(lang, model, field, fieldASTData, messages)
			}
		}
		messages = addResourceItemsToMessages(lang, filepath.Join(moduleDir, "resources"), messages)
		messages = addCodeToMessages(lang, moduleDir, messages)

		msgs := make([]po.Message, len(messages))
		i := 0
		for _, m := range messages {
			m.ExtractedComment = strings.TrimSuffix(m.ExtractedComment, "\n")
			msgs[i] = m
			i += 1
		}
		file := po.File{
			Messages: msgs,
			MimeHeader: po.Header{
				Language:                lang,
				ContentType:             "text/plain; charset=utf-8",
				ContentTransferEncoding: "8bit",
				MimeVersion:             "1.0",
			},
		}
		err = file.Save(fmt.Sprintf("%s/%s.po", i18nDir, lang))
		if err != nil {
			log.Panic("Error while saving PO file", "error", err)
		}
	}
}

// addCodeToMessages adds to the given messages map the translatable fields of the code
// defined in go files inside the given resourcesDir and sub directories.
// This extracts strings given as argument to T().
func addCodeToMessages(lang string, moduleDir string, messages map[messageRef]po.Message) map[messageRef]po.Message {
	fSet := token.NewFileSet()
	goFiles, err := filepath.Glob(fmt.Sprintf("%s/**.go", moduleDir))
	if err != nil {
		log.Panic("Unable to scan directory for go files", "moduleDir", moduleDir, "error", err)
	}
	for _, goFile := range goFiles {
		astFile, err := parser.ParseFile(fSet, goFile, nil, 0)
		if err != nil {
			log.Panic("Unable to parse file's AST", "file", goFile, "error", err)
		}
		ast.Inspect(astFile, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				fnctName, err := generate.ExtractFunctionName(node)
				if err != nil {
					return true
				}
				if fnctName != "T" {
					return true
				}
				strArg := strings.Trim(node.Args[0].(*ast.BasicLit).Value, "\"`")
				codeTrans := i18n.TranslateCode(lang, "", strArg)
				if codeTrans == strArg {
					codeTrans = ""
				}
				msgRef := messageRef{msgId: strArg}
				msg := getOrCreateMessage(messages, msgRef, codeTrans)
				msg.ExtractedComment += "code:\n"
				messages[msgRef] = msg
			}
			return true
		})
	}
	return messages
}

// addResourceItemsToMessages adds to the given messages map the translatable fields of the views
// defined in XML files inside the given resourcesDir
func addResourceItemsToMessages(lang string, resourcesDir string, messages map[messageRef]po.Message) map[messageRef]po.Message {
	xmlFiles, err := filepath.Glob(fmt.Sprintf("%s/*.xml", resourcesDir))
	if err != nil {
		log.Panic("Unable to scan directory for xml files", "dir", resourcesDir, "error", err)
	}
	for _, fileName := range xmlFiles {
		doc := etree.NewDocument()
		if err := doc.ReadFromFile(fileName); err != nil {
			log.Panic("Error loading XML data file", "file", fileName, "error", err)
		}
		viewColl := views.NewCollection()
		menuColl := make(map[string]*menus.Menu)
		actionColl := actions.NewCollection()
		for _, dataTag := range doc.FindElements("hexya/data") {
			for _, object := range dataTag.ChildElements() {
				switch object.Tag {
				case "view":
					viewColl.LoadFromEtree(object)
				case "menuitem":
					menus.AddMenuToMapFromEtree(object, menuColl)
				case "action":
					actionColl.LoadFromEtree(object)
				}
			}
		}
		for _, view := range viewColl.GetAll() {
			labels := view.TranslatableStrings()
			for _, label := range labels {
				messages = updateMessagesWithResourceTranslation(lang, view.ID, label.Value, messages)
			}
			// TODO add text data
		}
		for _, menu := range menuColl {
			messages = updateMessagesWithResourceTranslation(lang, menu.ID, menu.Name, messages)
		}
		for _, action := range actionColl.GetAll() {
			messages = updateMessagesWithResourceTranslation(lang, action.ID, action.Name, messages)
		}
	}
	return messages
}

// updateMessagesWithResourceTranslation returns the message map updated with a message
// corresponding to the given ID and source
func updateMessagesWithResourceTranslation(lang, id, source string, messages map[messageRef]po.Message) map[messageRef]po.Message {
	nameTrans := i18n.TranslateResourceItem(lang, id, source)
	if nameTrans == source {
		nameTrans = ""
	}
	msgRef := messageRef{msgId: source}
	msg := getOrCreateMessage(messages, msgRef, nameTrans)
	msg.ExtractedComment += fmt.Sprintf("resource:%s\n", id)
	messages[msgRef] = msg
	return messages
}

// addSelectionToMessages adds to the given messages map the selections for the given model and field
func addSelectionToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[messageRef]po.Message) map[messageRef]po.Message {
	if len(fieldASTData.Selection) == 0 {
		return messages
	}
	selection := types.Selection(fieldASTData.Selection)
	selTranslated := i18n.TranslateFieldSelection(lang, model, field, selection)
	for k, v := range selection {
		transValue := selTranslated[k]
		if transValue == v {
			transValue = ""
		}
		msgRef := messageRef{msgId: v}
		msg := getOrCreateMessage(messages, msgRef, transValue)
		msg.ExtractedComment += fmt.Sprintf("selection:%s.%s\n", model, field)
		messages[msgRef] = msg
	}
	return messages
}

// getOrCreateMessage retrieves the message in messages at the msgRef key.
// If it does not exist, then it is created with the given value.
// If value is not empty and the original msg translation is empty, then
// it is updated with value.
func getOrCreateMessage(messages map[messageRef]po.Message, msgRef messageRef, value string) po.Message {
	msg, ok := messages[msgRef]
	if !ok {
		msg = po.Message{
			MsgId: msgRef.msgId,
		}
	}
	if msg.MsgStr == "" {
		msg.MsgStr = value
	}
	return msg
}

// addDescriptionToMessages adds to the given messages map the description translation for the given model and field
func addDescriptionToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[messageRef]po.Message) map[messageRef]po.Message {
	description := fieldASTData.Description
	if description == "" {
		description = strutils.Title(fieldASTData.Name)
	}
	descTranslated := i18n.TranslateFieldDescription(lang, model, field, "")
	msgRef := messageRef{msgId: description}
	msg := getOrCreateMessage(messages, msgRef, descTranslated)
	msg.ExtractedComment += fmt.Sprintf("field:%s.%s\n", model, field)
	messages[msgRef] = msg
	return messages
}

// addHelpToMessages adds to the given messages map the help translation for the given model and field
func addHelpToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[messageRef]po.Message) map[messageRef]po.Message {
	help := fieldASTData.Help
	if help == "" {
		return messages
	}
	helpTranslated := i18n.TranslateFieldHelp(lang, model, field, "")
	msgRef := messageRef{msgId: help}
	msg := getOrCreateMessage(messages, msgRef, helpTranslated)
	msg.ExtractedComment += fmt.Sprintf("help:%s.%s\n", model, field)
	messages[msgRef] = msg
	return messages
}

func init() {
	i18nUpdate.PersistentFlags().StringSliceP("languages", "l", []string{}, "Comma separated list of languages codes to load (ex: fr,de,es).")
	HexyaCmd.AddCommand(i18nCmd)
	i18nCmd.AddCommand(i18nUpdate)
}

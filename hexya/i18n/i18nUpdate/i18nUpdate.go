package translation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/hexya/tools/po"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"github.com/hexya-erp/hexya/hexya/views"
	"golang.org/x/tools/go/loader"
)

var poUpdateDatas map[string]poUpdateFunc

var poRuleSets map[string]*RuleSet

var log logging.Logger

type poUpdateFunc func(MessageMap, string, string, string) MessageMap

// RuleSet contains the rules defining the files targeted by the custom i18n-Update methods
type RuleSet struct {
	Inherit []*RuleSet
	Ruleset [][]string
}

// MessageMap is a map with all po-related informations
type MessageMap map[MessageRef]po.Message

// UpdatePOFiles creates or updates PO files of the module in the given
// dir with the data in the Translation registry.
// It is meant to be called from
// a Po updater start file which imports all the project's module.
func UpdatePOFiles(config map[string]interface{}) {
	moduleDir := config["moduleDir"].(string)
	modulePath := config["modulePath"].(string)
	langs := config["langs"].([]string)
	if strings.ToUpper(langs[0]) == "ALL" {
		langs = append(i18n.GetAllLanguageList(), langs[1:]...)
	}
	log = logging.GetLogger("i18nUpdate")
	i18nDir := filepath.Join(moduleDir, "i18n")
	server.LoadModuleTranslations(i18nDir, langs)
	conf := loader.Config{}
	conf.Import(modulePath)
	fmt.Print("Loading...")
	program, err := conf.Load()
	fmt.Println("Ok.")
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
		fmt.Printf("Generating language %s.", lang)
		messages := make(map[MessageRef]po.Message)
		for model, modelASTData := range modelsASTData {
			for field, fieldASTData := range modelASTData.Fields {
				messages = addDescriptionToMessages(lang, model, field, fieldASTData, messages)
				messages = addHelpToMessages(lang, model, field, fieldASTData, messages)
				messages = addSelectionToMessages(lang, model, field, fieldASTData, messages)
			}
		}
		fmt.Printf(".")
		messages = addResourceItemsToMessages(lang, filepath.Join(moduleDir, "resources"), messages)
		messages = addCodeToMessages(lang, moduleDir, messages)

		moduleName := filepath.Base(moduleDir)

		messages = executeCustomPoFuncs(messages, lang, moduleName)

		msgs := make([]po.Message, len(messages))
		i := 0
		for _, m := range messages {
			m.ExtractedComment = strings.TrimSuffix(m.ExtractedComment, "\n")
			msgs[i] = m
			i += 1
		}
		fmt.Printf(".")
		file := po.File{
			Messages: msgs,
			MimeHeader: po.Header{
				Language:                lang,
				ContentType:             "text/plain; charset=utf-8",
				ContentTransferEncoding: "8bit",
				MimeVersion:             "1.0",
			},
		}
		poFileName := lang + ".po"
		err = file.Save(filepath.Join(i18nDir, poFileName))
		if err != nil {
			i18nDir = filepath.Join(strings.TrimSuffix(generate.HexyaDir, generate.HexyaPath), i18nDir)
			err2 := file.Save(filepath.Join(i18nDir, poFileName))
			if err2 != nil {
				log.Panic("Error while saving PO file", "error", err, "error", err2)
			}
		}
		fmt.Printf(" Done!\n")
	}
}

// A MessageRef identifies unique messages
type MessageRef struct {
	MsgId   string
	msgCtxt string
}

// Dummy type for typesafe registering
type registerer int

func Register() registerer {
	return 0
}

// Func registers the given method for the given key in the poMethods bundle
func (r registerer) Func(key string, f poUpdateFunc) {
	if poUpdateDatas == nil {
		poUpdateDatas = make(map[string]poUpdateFunc)
	}
	poUpdateDatas[key] = f
}

// RuleSet registers the given RuleSet for the given key in the poRuleSet bundle
func (r registerer) RuleSet(key string, rules *RuleSet) {
	if poRuleSets == nil {
		poRuleSets = make(map[string]*RuleSet)
	}
	poRuleSets[key] = rules
}

// GetPoUpdateRuleSet returns the rule set for an usage outside the package
func GetPoUpdateRuleSet(key string) *RuleSet {
	return poRuleSets[key]
}

//followRule returns true if the given path follows a full rule
func followsRule(str string, set []string) bool {
	for _, ruleLine := range set {
		excludeMode := false
		if strings.HasPrefix(ruleLine, "!") {
			ruleLine = strings.TrimPrefix(ruleLine, "!")
			excludeMode = true
		}
		rx := regexp.MustCompile(ruleLine)
		if rx.MatchString(str) == excludeMode {
			return false
		}
	}
	return true
}

// followsRules returns true if the given path follows a full RuleSet
func followsRules(str string, set *RuleSet) bool {
	if set == nil {
		return true
	}
	followsInherit := false
	if set.Inherit == nil {
		followsInherit = true
	}
	for _, inherit := range set.Inherit {
		if followsRules(str, inherit) {
			followsInherit = true
			break
		}
	}
	if followsInherit {
		for _, rule := range set.Ruleset {
			if followsRule(str, rule) {
				return true
			}
		}
	}
	return false
}

// executeCustomPoFuncs executes all methods registered by Hexya modules
func executeCustomPoFuncs(messages MessageMap, lang, moduleName string) MessageMap {
	for key, val := range poUpdateDatas {
		if val != nil {
			path := filepath.Join(generate.HexyaDir, "hexya/server/static", moduleName)
			fi, err := os.Lstat(path)
			if err != nil {
				return messages
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				absPath, err2 := os.Readlink(path)
				if err2 != nil {
					return messages
				}
				filepath.Walk(absPath, func(path2 string, info os.FileInfo, err3 error) error {
					goDir := strings.TrimSuffix(generate.HexyaDir, generate.HexyaPath)
					path2 = strings.TrimPrefix(path2, goDir)
					if info != nil && !info.IsDir() && followsRules(path2, poRuleSets[key]) {
						messages = val(messages, lang, path, moduleName)
					}
					return err3
				})
			}
		}
	}
	return messages
}

// addCodeToMessages adds to the given messages map the translatable fields of the code
// defined in go files inside the given resourcesDir and sub directories.
// This extracts strings given as argument to T().
func addCodeToMessages(lang string, moduleDir string, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
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
				msgRef := MessageRef{MsgId: strArg}
				msg := GetOrCreateMessage(messages, msgRef, codeTrans)
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
func addResourceItemsToMessages(lang string, resourcesDir string, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
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
func updateMessagesWithResourceTranslation(lang, id, source string, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
	nameTrans := i18n.TranslateResourceItem(lang, id, source)
	if nameTrans == source {
		nameTrans = ""
	}
	msgRef := MessageRef{MsgId: source}
	msg := GetOrCreateMessage(messages, msgRef, nameTrans)
	msg.ExtractedComment += fmt.Sprintf("resource:%s\n", id)
	messages[msgRef] = msg
	return messages
}

// addSelectionToMessages adds to the given messages map the selections for the given model and field
func addSelectionToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
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
		msgRef := MessageRef{MsgId: v}
		msg := GetOrCreateMessage(messages, msgRef, transValue)
		msg.ExtractedComment += fmt.Sprintf("selection:%s.%s\n", model, field)
		messages[msgRef] = msg
	}
	return messages
}

// GetOrCreateMessage retrieves the message in messages at the msgRef key.
// If it does not exist, then it is created with the given value.
// If value is not empty and the original msg translation is empty, then
// it is updated with value.
func GetOrCreateMessage(messages map[MessageRef]po.Message, msgRef MessageRef, value string) po.Message {
	msg, ok := messages[msgRef]
	if !ok {
		msg = po.Message{
			MsgId: msgRef.MsgId,
		}
	}
	if msg.MsgStr == "" {
		msg.MsgStr = value
	}
	return msg
}

// addDescriptionToMessages adds to the given messages map the description translation for the given model and field
func addDescriptionToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
	description := fieldASTData.Description
	if description == "" {
		description = strutils.Title(fieldASTData.Name)
	}
	descTranslated := i18n.TranslateFieldDescription(lang, model, field, "")
	msgRef := MessageRef{MsgId: description}
	msg := GetOrCreateMessage(messages, msgRef, descTranslated)
	msg.ExtractedComment += fmt.Sprintf("field:%s.%s\n", model, field)
	messages[msgRef] = msg
	return messages
}

// addHelpToMessages adds to the given messages map the help translation for the given model and field
func addHelpToMessages(lang string, model string, field string, fieldASTData generate.FieldASTData, messages map[MessageRef]po.Message) map[MessageRef]po.Message {
	help := fieldASTData.Help
	if help == "" {
		return messages
	}
	helpTranslated := i18n.TranslateFieldHelp(lang, model, field, "")
	msgRef := MessageRef{MsgId: help}
	msg := GetOrCreateMessage(messages, msgRef, helpTranslated)
	msg.ExtractedComment += fmt.Sprintf("help:%s.%s\n", model, field)
	messages[msgRef] = msg
	return messages
}

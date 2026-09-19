// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package translations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexya-erp/hexya/src/i18n"
	"github.com/hexya-erp/hexya/src/tools/generate"
	"github.com/hexya-erp/hexya/src/tools/logging"
	"github.com/hexya-erp/hexya/src/tools/po"
)

var (
	testModuleDir   = filepath.Join("testdata", "testmodule")
	testResourceDir = filepath.Join(testModuleDir, "resources")
)

func TestMain(m *testing.M) {
	// log is only set by UpdatePOFiles, so we must set it here for the
	// functions which may log.
	log = logging.GetLogger("i18nUpdate")
	i18n.LoadPOFile(filepath.Join("testdata", "fr.po"))
	os.Exit(m.Run())
}

func TestFollowsRule(t *testing.T) {
	t.Run("Empty rule should always match", func(t *testing.T) {
		assert.True(t, followsRule("foo/bar.xml", []string{}))
	})
	t.Run("Simple inclusion rules", func(t *testing.T) {
		assert.True(t, followsRule("foo/bar.xml", []string{`\.xml$`}))
		assert.False(t, followsRule("foo/bar.go", []string{`\.xml$`}))
	})
	t.Run("All lines of a rule must match", func(t *testing.T) {
		assert.True(t, followsRule("foo/bar.xml", []string{`^foo/`, `\.xml$`}))
		assert.False(t, followsRule("baz/bar.xml", []string{`^foo/`, `\.xml$`}))
	})
	t.Run("Exclusion rules", func(t *testing.T) {
		assert.True(t, followsRule("foo/bar.xml", []string{`\.xml$`, `!^baz/`}))
		assert.False(t, followsRule("foo/bar.xml", []string{`\.xml$`, `!^foo/`}))
	})
}

func TestFollowsRules(t *testing.T) {
	t.Run("A nil rule set should always match", func(t *testing.T) {
		assert.True(t, followsRules("foo/bar.xml", nil))
	})
	t.Run("A rule set without inherit", func(t *testing.T) {
		set := &RuleSet{Rules: [][]string{{`\.xml$`}, {`\.csv$`}}}
		assert.True(t, followsRules("foo/bar.xml", set))
		assert.True(t, followsRules("foo/bar.csv", set))
		assert.False(t, followsRules("foo/bar.go", set))
	})
	t.Run("A rule set with no rule at all should not match", func(t *testing.T) {
		assert.False(t, followsRules("foo/bar.xml", &RuleSet{}))
	})
	t.Run("A rule set with inherit", func(t *testing.T) {
		parent := &RuleSet{Rules: [][]string{{`^foo/`}}}
		set := &RuleSet{
			Inherit: []*RuleSet{parent},
			Rules:   [][]string{{`\.xml$`}},
		}
		assert.True(t, followsRules("foo/bar.xml", set))
		// Does not follow the inherited rule set
		assert.False(t, followsRules("baz/bar.xml", set))
		// Follows the inherited rule set but not its own rules
		assert.False(t, followsRules("foo/bar.go", set))
	})
}

func TestRegistries(t *testing.T) {
	t.Run("Registering and retrieving rule sets", func(t *testing.T) {
		poRuleSets = nil
		assert.Nil(t, GetPoUpdateRuleSet("myKey"))
		ruleSet := &RuleSet{Rules: [][]string{{`\.xml$`}}}
		RegisterRuleSet("myKey", ruleSet)
		assert.Equal(t, ruleSet, GetPoUpdateRuleSet("myKey"))
		// Registering a second time should not reset the bundle
		RegisterRuleSet("myOtherKey", &RuleSet{})
		assert.NotNil(t, GetPoUpdateRuleSet("myKey"))
		assert.NotNil(t, GetPoUpdateRuleSet("myOtherKey"))
		poRuleSets = nil
	})
	t.Run("Registering po update functions", func(t *testing.T) {
		poUpdateFuncs = nil
		fnct := func(messages MessageMap, _, _, _ string) MessageMap { return messages }
		RegisterFunc("myKey", fnct)
		assert.Len(t, poUpdateFuncs, 1)
		assert.Contains(t, poUpdateFuncs, "myKey")
		RegisterFunc("myOtherKey", nil)
		assert.Len(t, poUpdateFuncs, 2)
		poUpdateFuncs = nil
	})
}

func TestGetOrCreateMessage(t *testing.T) {
	t.Run("Creating a new message", func(t *testing.T) {
		messages := make(MessageMap)
		msg := GetOrCreateMessage(messages, MessageRef{MsgId: "Hello"}, "Bonjour")
		assert.Equal(t, "Hello", msg.MsgId)
		assert.Equal(t, "Bonjour", msg.MsgStr)
	})
	t.Run("Retrieving an existing translated message should not override it", func(t *testing.T) {
		msgRef := MessageRef{MsgId: "Hello"}
		messages := MessageMap{msgRef: po.Message{MsgId: "Hello", MsgStr: "Salut"}}
		msg := GetOrCreateMessage(messages, msgRef, "Bonjour")
		assert.Equal(t, "Salut", msg.MsgStr)
	})
	t.Run("Retrieving an existing untranslated message should set the value", func(t *testing.T) {
		msgRef := MessageRef{MsgId: "Hello"}
		messages := MessageMap{msgRef: po.Message{MsgId: "Hello"}}
		msg := GetOrCreateMessage(messages, msgRef, "Bonjour")
		assert.Equal(t, "Bonjour", msg.MsgStr)
	})
}

func TestAddDescriptionToMessages(t *testing.T) {
	t.Run("Field with an explicit description", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addDescriptionToMessages("fr", "TestModel", "Name",
			generate.FieldASTData{Name: "Name", Description: "Name"}, messages)
		assert.Len(t, messages, 1)
		msg := messages[MessageRef{MsgId: "Name"}]
		assert.Equal(t, "Name", msg.MsgId)
		assert.Equal(t, "Nom", msg.MsgStr)
		assert.Equal(t, "field:TestModel.Name\n", msg.ExtractedComment)
	})
	t.Run("Field without description should use its title", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addDescriptionToMessages("fr", "TestModel", "UserName",
			generate.FieldASTData{Name: "UserName"}, messages)
		assert.Len(t, messages, 1)
		assert.Contains(t, messages, MessageRef{MsgId: "User Name"})
		assert.Equal(t, "", messages[MessageRef{MsgId: "User Name"}].MsgStr)
	})
	t.Run("Two fields with the same description", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addDescriptionToMessages("fr", "TestModel", "Name",
			generate.FieldASTData{Name: "Name", Description: "Name"}, messages)
		messages = addDescriptionToMessages("fr", "OtherModel", "Name",
			generate.FieldASTData{Name: "Name", Description: "Name"}, messages)
		assert.Len(t, messages, 1)
		msg := messages[MessageRef{MsgId: "Name"}]
		assert.Equal(t, "field:TestModel.Name\nfield:OtherModel.Name\n", msg.ExtractedComment)
	})
}

func TestAddHelpToMessages(t *testing.T) {
	t.Run("Field without help should be ignored", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addHelpToMessages("fr", "TestModel", "Name", generate.FieldASTData{Name: "Name"}, messages)
		assert.Len(t, messages, 0)
	})
	t.Run("Field with help", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addHelpToMessages("fr", "TestModel", "Name",
			generate.FieldASTData{Name: "Name", Help: "The name of the record"}, messages)
		assert.Len(t, messages, 1)
		msg := messages[MessageRef{MsgId: "The name of the record"}]
		assert.Equal(t, "Le nom de l'enregistrement", msg.MsgStr)
		assert.Equal(t, "help:TestModel.Name\n", msg.ExtractedComment)
	})
}

func TestAddSelectionToMessages(t *testing.T) {
	t.Run("Field without selection should be ignored", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addSelectionToMessages("fr", "TestModel", "State", generate.FieldASTData{Name: "State"}, messages)
		assert.Len(t, messages, 0)
	})
	t.Run("Field with selection", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addSelectionToMessages("fr", "TestModel", "State", generate.FieldASTData{
			Name:      "State",
			Selection: map[string]string{"active": "Active", "inactive": "Inactive"},
		}, messages)
		assert.Len(t, messages, 2)
		msg := messages[MessageRef{MsgId: "Active"}]
		assert.Equal(t, "Actif", msg.MsgStr)
		assert.Equal(t, "selection:TestModel.State\n", msg.ExtractedComment)
		// Untranslated selection items have an empty MsgStr
		assert.Equal(t, "", messages[MessageRef{MsgId: "Inactive"}].MsgStr)
	})
}

func TestUpdateMessagesWithResourceTranslation(t *testing.T) {
	t.Run("Translated resource", func(t *testing.T) {
		messages := make(MessageMap)
		messages = updateMessagesWithResourceTranslation("fr", "test_view_id", "Profile Data", messages)
		assert.Len(t, messages, 1)
		msg := messages[MessageRef{MsgId: "Profile Data"}]
		assert.Equal(t, "Données du profil", msg.MsgStr)
		assert.Equal(t, "resource:test_view_id\n", msg.ExtractedComment)
	})
	t.Run("Untranslated resource", func(t *testing.T) {
		messages := make(MessageMap)
		messages = updateMessagesWithResourceTranslation("fr", "unknown_id", "Unknown Data", messages)
		assert.Len(t, messages, 1)
		assert.Equal(t, "", messages[MessageRef{MsgId: "Unknown Data"}].MsgStr)
	})
}

func TestAddResourceItemsToMessages(t *testing.T) {
	t.Run("Unknown directory should not add anything", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addResourceItemsToMessages("fr", filepath.Join("testdata", "unknown"), messages)
		assert.Len(t, messages, 0)
	})
	t.Run("Loading views, menus and actions", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addResourceItemsToMessages("fr", testResourceDir, messages)
		assert.Len(t, messages, 3)
		assert.Equal(t, "Données du profil", messages[MessageRef{MsgId: "Profile Data"}].MsgStr)
		assert.Equal(t, "resource:test_view_id\n", messages[MessageRef{MsgId: "Profile Data"}].ExtractedComment)
		assert.Equal(t, "Menu de test", messages[MessageRef{MsgId: "Test Menu"}].MsgStr)
		assert.Equal(t, "resource:test_menu_id\n", messages[MessageRef{MsgId: "Test Menu"}].ExtractedComment)
		assert.Equal(t, "Action de test", messages[MessageRef{MsgId: "Test Action"}].MsgStr)
		assert.Equal(t, "resource:test_action_id\n", messages[MessageRef{MsgId: "Test Action"}].ExtractedComment)
	})
}

func TestAddCodeToMessages(t *testing.T) {
	t.Run("Directory without go files should not add anything", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addCodeToMessages("fr", testResourceDir, messages)
		assert.Len(t, messages, 0)
	})
	t.Run("Extracting strings given to T()", func(t *testing.T) {
		messages := make(MessageMap)
		messages = addCodeToMessages("fr", testModuleDir, messages)
		assert.Len(t, messages, 2)
		assert.Equal(t, "Bonjour le monde", messages[MessageRef{MsgId: "Hello World"}].MsgStr)
		assert.Equal(t, "code:\n", messages[MessageRef{MsgId: "Hello World"}].ExtractedComment)
		assert.Contains(t, messages, MessageRef{MsgId: "Goodbye"})
		assert.Equal(t, "", messages[MessageRef{MsgId: "Goodbye"}].MsgStr)
		// Strings given to other functions are not extracted
		assert.NotContains(t, messages, MessageRef{MsgId: "Not translated"})
	})
}

func TestExecuteCustomPoFuncs(t *testing.T) {
	defer func() {
		poUpdateFuncs = nil
		poRuleSets = nil
	}()
	t.Run("Without any registered func", func(t *testing.T) {
		poUpdateFuncs = nil
		poRuleSets = nil
		messages := make(MessageMap)
		messages = executeCustomPoFuncs("fr", testModuleDir, messages)
		assert.Len(t, messages, 0)
	})
	t.Run("With a registered func and no rule set", func(t *testing.T) {
		poUpdateFuncs = nil
		poRuleSets = nil
		var visited []string
		var modules []string
		RegisterFunc("all", func(messages MessageMap, lang, path, module string) MessageMap {
			visited = append(visited, filepath.Base(path))
			modules = append(modules, module)
			messages[MessageRef{MsgId: filepath.Base(path)}] = po.Message{MsgId: filepath.Base(path), MsgStr: lang}
			return messages
		})
		messages := executeCustomPoFuncs("fr", testModuleDir, make(MessageMap))
		assert.Contains(t, visited, "testmodule.go")
		assert.Contains(t, visited, "views.xml")
		assert.Contains(t, modules, "testmodule")
		assert.Equal(t, "fr", messages[MessageRef{MsgId: "views.xml"}].MsgStr)
	})
	t.Run("With a rule set restricting the files", func(t *testing.T) {
		poUpdateFuncs = nil
		poRuleSets = nil
		var visited []string
		RegisterFunc("xmlOnly", func(messages MessageMap, _, path, _ string) MessageMap {
			visited = append(visited, filepath.Base(path))
			return messages
		})
		RegisterRuleSet("xmlOnly", &RuleSet{Rules: [][]string{{`\.xml$`}}})
		executeCustomPoFuncs("fr", testModuleDir, make(MessageMap))
		assert.Equal(t, []string{"views.xml"}, visited)
	})
	t.Run("A nil func should be skipped", func(t *testing.T) {
		poUpdateFuncs = nil
		poRuleSets = nil
		RegisterFunc("nilFunc", nil)
		assert.NotPanics(t, func() { executeCustomPoFuncs("fr", testModuleDir, make(MessageMap)) })
	})
}

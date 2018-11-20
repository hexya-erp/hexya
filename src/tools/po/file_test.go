// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package po

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	testPoFile = "testdata/test_file.po"
)

func TestReadPOFile(t *testing.T) {
	po, err := Load(testPoFile)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&po.MimeHeader, &testFile.MimeHeader) {
		t.Fatalf("Expected:\n---------\n%#v\n\nGot:\n----\n%#v", &testFile.MimeHeader, &po.MimeHeader)
	}
	if len(po.Messages) != len(testFile.Messages) {
		t.Fatal("size not equal")
	}
	for k, v0 := range po.Messages {
		if v1 := testFile.Messages[k]; !reflect.DeepEqual(&v0, &v1) {
			t.Fatalf("%d: \nExpected:\n---------\n%#v\n\nGot:\n----\n%#v", k, v1, v0)
		}
	}
}

func TestWritePOFile(t *testing.T) {
	fileName := filepath.Join(os.TempDir(), "testfile.po")
	testFile.Save(fileName)
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal("Unable to reload saved PO file")
	}
	if string(fileData) != resultTestFileString {
		t.Fatalf("\nExpected:\n---------\n%#v\n\nGot:\n----\n%#v", resultTestFileString, string(fileData))
	}
}

var testFile = &File{
	MimeHeader: Header{
		Comment:                 Comment{StartLine: 1, TranslatorComment: "SOME DESCRIPTIVE TITLE.\nCopyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER\nThis file is distributed under the same license as the PACKAGE package.\nFIRST AUTHOR <EMAIL@ADDRESS>, YEAR.\n"},
		ProjectIdVersion:        "Poedit 1.5",
		ReportMsgidBugsTo:       "poedit@googlegroups.com",
		POTCreationDate:         "2012-07-30 10:34+0200",
		PORevisionDate:          "2013-02-24 21:00+0800",
		LastTranslator:          "Christopher Meng <trans@cicku.me>",
		Language:                "zh_CN",
		LanguageTeam:            "",
		MimeVersion:             "1.0",
		ContentType:             "text/plain; charset=UTF-8",
		ContentTransferEncoding: "8bit",
		PluralForms:             "nplurals=1; plural=0;",
		XGenerator:              "Poedit 1.5.5",
	},
	Messages: []Message{
		{
			Comment: Comment{StartLine: 21, ReferenceFile: []string{"../src/edframe.cpp"}, ReferenceLine: []int{2060}},
			MsgId:   " (modified)",
			MsgStr:  " (已修改)",
		},
		{
			Comment: Comment{StartLine: 25, ExtractedComment: "TRANSLATORS: This is version information in about dialog, it is followed\nby version number when used (wxWidgets 2.8)", ReferenceFile: []string{"../src/edframe.cpp"}, ReferenceLine: []int{2431}},
			MsgId:   " Version \\",
			MsgStr:  " 版本 \\",
		},
		{
			Comment:      Comment{StartLine: 31, ReferenceFile: []string{"../src/edframe.cpp"}, ReferenceLine: []int{1367}, Flags: []string{"c-format"}},
			MsgId:        "%d issue with the translation found.",
			MsgIdPlural:  "%d issues with the translation found.",
			MsgStrPlural: []string{"在翻译中发现了 %d 个问题。"},
		},
		{
			Comment:      Comment{StartLine: 37, ReferenceFile: []string{"../src/edframe.cpp"}, ReferenceLine: []int{2024}, Flags: []string{"c-format"}},
			MsgId:        "%i %% translated\n\t %i string",
			MsgIdPlural:  "%i %% translated\n\t %i strings",
			MsgStrPlural: []string{"%i%% 已翻译\n\t %i 个字串"},
		},
		{
			Comment:      Comment{StartLine: 43, ReferenceFile: []string{"../src/edframe.cpp"}, ReferenceLine: []int{2029}, Flags: []string{"c-format"}},
			MsgContext:   "some-context-string",
			MsgId:        "%i %% translated, %i string (%s)",
			MsgIdPlural:  "%i %% translated, %i strings (%s)",
			MsgStrPlural: []string{"%i%% 已翻译，%i 个字串 (%s)"}},
	},
}

var resultTestFileString string = `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#` + " " + `
msgid ""
msgstr ""
"Project-Id-Version: Poedit 1.5\n"
"Report-Msgid-Bugs-To: poedit@googlegroups.com\n"
"POT-Creation-Date: 2012-07-30 10:34+0200\n"
"PO-Revision-Date: 2013-02-24 21:00+0800\n"
"Last-Translator: Christopher Meng <trans@cicku.me>\n"
"Language-Team: \n"
"Language: zh_CN\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=1; plural=0;\n"
"X-Generator: Poedit 1.5.5\n"

#: ../src/edframe.cpp:2060
msgid " (modified)"
msgstr " (已修改)"

#. TRANSLATORS: This is version information in about dialog, it is followed
#. by version number when used (wxWidgets 2.8)
#: ../src/edframe.cpp:2431
msgid " Version \\"
msgstr " 版本 \\"

#: ../src/edframe.cpp:1367
#, c-format
msgid "%d issue with the translation found."
msgid_plural "%d issues with the translation found."
msgstr[0] "在翻译中发现了 %d 个问题。"

#: ../src/edframe.cpp:2024
#, c-format
msgid ""
"%i %% translated\n"
"\t %i string"
msgid_plural ""
"%i %% translated\n"
"\t %i strings"
msgstr[0] ""
"%i%% 已翻译\n"
"\t %i 个字串"

#: ../src/edframe.cpp:2029
#, c-format
msgctxt "some-context-string"
msgid "%i %% translated, %i string (%s)"
msgid_plural "%i %% translated, %i strings (%s)"
msgstr[0] "%i%% 已翻译，%i 个字串 (%s)"

`

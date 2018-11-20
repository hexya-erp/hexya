// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package po

import (
	"reflect"
	"testing"
)

func TestPoEntry(t *testing.T) {
	var entry Message
	for i, testEntry := range testPoEntries {
		if err := entry.readPoEntry(newLineReader(testEntry.poEntryString)); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(&entry, &testEntry.message) {
			t.Fatalf("%d: \nExpected:\n---------\n%#v\n\nGot:\n----\n%#v", i, testEntry.message, entry)
		}
	}
}

type messageTestCase struct {
	poEntryString string
	message       Message
}

var testPoEntries = []messageTestCase{
	{
		poEntryString: `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
msgid ""
msgstr ""
"Project-Id-Version: 项目名称\n"
"Report-Msgid-Bugs-To: \n"
"POT-Creation-Date: 2011-12-12 20:03+0000\n"
"PO-Revision-Date: 2013-12-02 17:05+0800\n"
"Last-Translator: chai2010 <chaishushan@gmail.com>\n"
"Language-Team: chai2010(团队) <chaishushan@gmail.com>\n"
"Language: 中文\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"X-Generator: Poedit 1.5.7\n"
"X-Poedit-SourceCharset: UTF-8\n"
`,
		message: Message{
			Comment: Comment{
				StartLine: 1,
				TranslatorComment: `SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
This file is distributed under the same license as the PACKAGE package.
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
`,
			},
			MsgStr: `Project-Id-Version: 项目名称
Report-Msgid-Bugs-To: 
POT-Creation-Date: 2011-12-12 20:03+0000
PO-Revision-Date: 2013-12-02 17:05+0800
Last-Translator: chai2010 <chaishushan@gmail.com>
Language-Team: chai2010(团队) <chaishushan@gmail.com>
Language: 中文
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit
X-Generator: Poedit 1.5.7
X-Poedit-SourceCharset: UTF-8
`,
		},
	},
}

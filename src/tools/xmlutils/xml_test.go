// Copyright 2018 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package xmlutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beevik/etree"
)

func TestConcatXML(t *testing.T) {
	t.Run("Testing XML Concatenation", func(t *testing.T) {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xml2.xml", "testdata/xml3.xml"})
		assert.EqualValues(t, string(res), `<rootTag>
    <firstTag>
        Foo
    </firstTag>
    <secondTag>
        Bar
    </secondTag>

    <data>
Lorem Ipsum
    </data>
    <data2/>
</rootTag>`)
		assert.EqualValues(t, fmt.Sprintf("%x", sha), "e8965a6008bac9638d86a804d78ab8f2ca30a06d")
		assert.Nil(t, err)
	})
	t.Run("Non existent file should fail", func(t *testing.T) {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xml-not-exists.xml"})
		assert.Empty(t, res)
		assert.EqualValues(t, fmt.Sprintf("%x", sha), "0000000000000000000000000000000000000000")
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "unable to open XML file testdata/xml-not-exists.xml: open testdata/xml-not-exists.xml: no such file or directory")
	})
	t.Run("Invalid XML input should fail", func(t *testing.T) {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xmlfail.xml"})
		assert.Empty(t, res)
		assert.EqualValues(t, fmt.Sprintf("%x", sha), "0000000000000000000000000000000000000000")
		assert.NotNil(t, err)
		assert.EqualValues(t, err.Error(), "unable to parse XML file testdata/xmlfail.xml: EOF")
	})
}

var (
	baseXML = `
<form>
	<h1><field name="Name"/></h1>
	<group name="position_info">
		<field name="Function"/>
	</group>
	<group name="contact_data">
		<field name="Email"/>
	</group>
</form>
`
	baseTemplate = `
	<script type="text/javascript" src="/path/to/my/src.js"> </script>
	<script type="text/javascript" src="/path/to/my/other/src.js"> </script>
`
	specs = `
<group name="position_info" position="inside">
	<field name="CompanyName"/>
</group>
<xpath expr="//field[@name='Email']" position="after">
	<field name="Phone"/>
</xpath>
<group name="contact_data" position="before">
	<group>
		<field name="Address"/>
	</group>
	<hr/>
</group>
<h1 position="replace">
	<h2><field name="Name"/></h2>
</h1>
<xpath expr="//field[@name='Address']/.." position="attributes">
	<attribute name="name">address</attribute>
	<attribute name="string">Address</attribute>
</xpath>
`
	templateAppendSpec = `
<xpath expr="." position="inside">
	<script type="text/javascript" src="/path/to/third/src.js"> </script>
</xpath>
`
	notASpec = `
<foo>
	Bar
</foo>
`
	noParentSpec = `
<field name="noSuchField" position="after">
	<do foo="bar"/>
</field>
`
	noPositionSpec = `
<field name="Email">
	<field name="Something"/>
</field>
`
)

func TestApplyExtensions(t *testing.T) {
	t.Run("Testing ApplyExtensions", func(t *testing.T) {
		t.Run("Correct XML Extension spec", func(t *testing.T) {
			baseElem, _ := XMLToDocument(baseXML)
			specDoc := etree.NewDocument()
			specDoc.ReadFromString(specs)
			res, err := ApplyExtensions(baseElem, specDoc)
			assert.Nil(t, err)
			xml, _ := DocumentToXML(res)
			assert.EqualValues(t, string(xml), `<form>
	<h2>
		<field name="Name"/>
	</h2>
	<group name="position_info">
		<field name="Function"/>
		<field name="CompanyName"/>
	</group>
	<group name="address" string="Address">
		<field name="Address"/>
	</group>
	<hr/>
	<group name="contact_data">
		<field name="Email"/>
		<field name="Phone"/>
	</group>
</form>
`)
		})
		t.Run("Template spec", func(t *testing.T) {
			specDoc := etree.NewDocument()
			baseTmplDoc, _ := XMLToDocument(baseTemplate)
			specDoc.ReadFromString(templateAppendSpec)
			res, err := ApplyExtensions(baseTmplDoc, specDoc)
			assert.Nil(t, err)
			xml, _ := DocumentToXMLNoIndent(res)
			assert.EqualValues(t, string(xml), `
	<script type="text/javascript" src="/path/to/my/src.js"> </script>
	<script type="text/javascript" src="/path/to/my/other/src.js"> </script>
	<script type="text/javascript" src="/path/to/third/src.js"> </script>
`)
		})
		t.Run("XML which is not a spec should fail", func(t *testing.T) {
			baseElem, _ := XMLToDocument(baseXML)
			specDoc := etree.NewDocument()
			specDoc.ReadFromString(notASpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), `error in spec <foo>
	Bar
</foo>
: invalid view inherit spec`)
			assert.Nil(t, res)
		})
		t.Run("Specs for unknown parent should fail", func(t *testing.T) {
			baseElem, _ := XMLToDocument(baseXML)
			specDoc := etree.NewDocument()
			specDoc.ReadFromString(noParentSpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), "node not found in parent view: //field[@name='noSuchField']")
			assert.Nil(t, res)
		})
		t.Run("Specs without position attribute should fail", func(t *testing.T) {
			baseElem, _ := XMLToDocument(baseXML)
			specDoc := etree.NewDocument()
			specDoc.ReadFromString(noPositionSpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			assert.NotNil(t, err)
			assert.EqualValues(t, err.Error(), `spec should include 'position' attribute : <field name="Email">
	<field name="Something"/>
</field>
`)
			assert.Nil(t, res)
		})
	})
}

func TestHasParentTag(t *testing.T) {
	t.Run("Checking parent tag", func(t *testing.T) {
		baseElem, _ := XMLToElement(baseXML)
		field := baseElem.FindElement("//field[@name='Function']")
		assert.True(t, HasParentTag(field, "group"))
		assert.True(t, HasParentTag(field, "form"))
		assert.False(t, HasParentTag(field, "h1"))
		assert.False(t, HasParentTag(field, "field"))
	})
}

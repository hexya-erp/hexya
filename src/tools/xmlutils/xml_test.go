// Copyright 2018 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package xmlutils

import (
	"fmt"
	"testing"

	"github.com/beevik/etree"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConcatXML(t *testing.T) {
	Convey("Testing XML Concatenation", t, func() {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xml2.xml", "testdata/xml3.xml"})
		So(string(res), ShouldEqual, `<rootTag>
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
		So(fmt.Sprintf("%x", sha), ShouldEqual, "e8965a6008bac9638d86a804d78ab8f2ca30a06d")
		So(err, ShouldBeNil)
	})
	Convey("Non existent file should fail", t, func() {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xml-not-exists.xml"})
		So(res, ShouldBeEmpty)
		So(fmt.Sprintf("%x", sha), ShouldEqual, "0000000000000000000000000000000000000000")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unable to open XML file testdata/xml-not-exists.xml: open testdata/xml-not-exists.xml: no such file or directory")
	})
	Convey("Invalid XML input should fail", t, func() {
		res, sha, err := ConcatXML([]string{"testdata/xml1.xml", "testdata/xmlfail.xml"})
		So(res, ShouldBeEmpty)
		So(fmt.Sprintf("%x", sha), ShouldEqual, "0000000000000000000000000000000000000000")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unable to parse XML file testdata/xmlfail.xml: EOF")
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
	Convey("Testing ApplyExtensions", t, func() {
		baseElem, _ := XMLToElement(baseXML)
		specDoc := etree.NewDocument()
		Convey("Correct XML Extension spec", func() {
			specDoc.ReadFromString(specs)
			res, err := ApplyExtensions(baseElem, specDoc)
			So(err, ShouldBeNil)
			xml, _ := ElementToXML(res)
			So(string(xml), ShouldEqual, `<form>
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
		Convey("XML which is not a spec should fail", func() {
			specDoc.ReadFromString(notASpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `error in spec <foo>
	Bar
</foo>
: invalid view inherit spec`)
			So(res, ShouldBeNil)
		})
		Convey("Specs for unknown parent should fail", func() {
			specDoc.ReadFromString(noParentSpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node not found in parent view: //field[@name='noSuchField']")
			So(res, ShouldBeNil)
		})
		Convey("Specs without position attribute should fail", func() {
			specDoc.ReadFromString(noPositionSpec)
			res, err := ApplyExtensions(baseElem, specDoc)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `spec should include 'position' attribute : <field name="Email">
	<field name="Something"/>
</field>
`)
			So(res, ShouldBeNil)
		})
	})
}

func TestHasParentTag(t *testing.T) {
	Convey("Checking parent tag", t, func() {
		baseElem, _ := XMLToElement(baseXML)
		field := baseElem.FindElement("//field[@name='Function']")
		So(HasParentTag(field, "group"), ShouldBeTrue)
		So(HasParentTag(field, "form"), ShouldBeTrue)
		So(HasParentTag(field, "h1"), ShouldBeFalse)
		So(HasParentTag(field, "field"), ShouldBeFalse)
	})
}

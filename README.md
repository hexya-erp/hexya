[![Build Status](https://travis-ci.com/hexya-erp/hexya.svg?branch=master)](https://travis-ci.com/hexya-erp/hexya)
[![Go Report Card](https://goreportcard.com/badge/hexya-erp/hexya)](https://goreportcard.com/report/hexya-erp/hexya)
[![codecov](https://codecov.io/gh/hexya-erp/hexya/branch/master/graph/badge.svg)](https://codecov.io/gh/hexya-erp/hexya)
[![godoc reference](https://godoc.org/github.com/hexya-erp/hexya?status.png)](https://godoc.org/github.com/hexya-erp/hexya)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Hexya

Hexya is an open source ERP and a business application development framework
written in Go.

This repository houses the business application development framework.
The ERP is built by integrating modules of the [Hexya Addons Project](https://github.com/hexya-addons)

## Features of the framework

The Hexya framework is designed to develop business applications quickly and safely.
It includes all needed components in a very opinionated way.

The examples below are here to give you a first idea of Hexya. 

Head to the `/doc` directory and especially our [Tutorial](./doc/tutorial.adoc) if you want to start developing your business application with Hexya.

### ORM

Hexya includes a full-featured type safe ORM, including a type safe query builder.

Declare a model and add some fields
```go
var fields_User = map[string]models.FieldDefinition{
    "Name": fields.Char{String: "Name", Help: "The user's username", Unique: true,
        NoCopy: true, OnChange: h.User().Methods().OnChangeName()},
    "Email":    fields.Char{Help: "The user's email address", Size: 100, Index: true},
    "Password": fields.Char{},
    "IsStaff":  fields.Boolean{String: "Is a Staff Member", 
        Help: "Set to true if this user is a member of staff"},
}

func init() {
	models.NewModel("User")
	h.User().AddFields(fields_User)
}
```

Use the ORM to create a record in the database with type-safe data
```go
newUser := h.User().Create(env, h.User().NewData().
	SetName("John").
	SetEmail("john@example.com").
	SetIsStaff(true))
```

Search the database using the type-safe query builder and update records directly
```go
myUsers := h.User().Search(env,
	q.User().Name().Contains("John").
		And().Email().NotEquals("contact@example.com"))
for _, myUser := range myUsers.Records() {
    if myUser.IsStaff() {
        myUser.SetEmail("contact@example.com")
    }	
}
```

Add methods to the models
```go
// GetEmail returns the Email of the user with the given name
func user_GetEmail(rs m.UserSet, name string) string {
    user := h.User().Search(env, q.User().Name().Equals("John")).Limit(1)
    user.Sanitize()     // Call other methods of the model
    return user.Email() // If user is empty, then Email() will return the empty string
}

func init() {
    h.User().NewMethod("GetEmail", user_GetEmail)
}
```

### Views

Define views of different types using a simple XML view definition and let the framework do the rendering:

```xml
<view id="base_view_users_form_simple_modif" model="User" priority="18">
    <form string="Users">
        <field name="image" readonly="0" widget='image'
               options='{"preview_image": "image_small"}'/>
        <h1>
            <field name="name" readonly="1"/>
        </h1>
        <button name="preference_change_password" type="object" string="Change password"/>
        <group name="preferences" col="4">
            <field name="lang" readonly="0"/>
            <field name="tz" widget="timezone_mismatch" options="{'tz_offset_field': 'tz_offset'}"
                   readonly="0"/>
            <field name="tz_offset" invisible="1"/>
            <field name="company_id" options="{'no_create': True}" readonly="0"
                   groups="base_group_multi_company"/>
        </group>
        <group string="Email Preferences">
            <field name="email" widget="email" readonly="0"/>
            <field name="signature" readonly="0"/>
        </group>
        <footer>
            <button name="preference_save" type="object" string="Save" class="btn-primary"/>
            <button name="preference_cancel" string="Cancel" special="cancel" class="btn-default"/>
        </footer>
    </form>
</view>
```

### Controllers

Most of the time, you do not need to declare controllers in Hexya. 
Instead, declare an "Action" with the views you want and a menu to access it.
The framework will take care of the UI including rendering views, navigation, CRUD, etc.

```xml
<action id="base_action_res_users" 
        type="ir.actions.act_window" 
        name="Users" 
        model="User"
        view_id="base_view_users_tree" 
        search_view_id="base_view_users_search" 
        view_mode="tree,form,calendar"/>

<menuitem id="base_menu_action_users" 
          name="Users" 
          sequence="1" 
          action="base_action_res_users"
          parent="base_menu_users"/>
```

### Iterative Definition and Modularity

Each part of the Hexya Framework is modular and follow the Iterative Definition concept.

This means that an object (for example a model class) can be defined in a module and then extended in place by dependent modules.
So any subsequent modification will be made on the original model and will be available for the whole application.

This makes it possible to customize the application by creating a new module with the new features without forking and still benefiting from upstream updates.

Example on models:
```go
package A

var fields_User = map[string]models.FieldDefinition{
    "Name": fields.Char{String: "Name", Help: "The user's username", Unique: true,
        NoCopy: true, OnChange: h.User().Methods().OnChangeName()},
    "Email":    fields.Char{Help: "The user's email address", Size: 100, Index: true},
    "Password": fields.Char{},
    "IsStaff":  fields.Boolean{String: "Is a Staff Member", 
        Help: "Set to true if this user is a member of staff"},
}

func init() {
    models.NewModel("User")
    h.User().AddFields(fields_User)
}
```
```go
package B

var fields_User = map[string]models.FieldDefinition{
    "Size": models.Float{},
}

func init() {
    h.User().AddFields(fields_User)
}
```
```go
// Anywhere else
newUser := h.User().Create(env, h.User().NewData().
	SetName("John").
	SetEmail("john@example.com").
	SetSize(185.7))
fmt.Println(newUser.Name())
// output : John
fmt.Println(newUser.Size())
// output : 185.7
```

Model methods can be extended too:

```go
package A

// GetEmail returns the Email of the user with the given name
func user_GetEmail(rs m.UserSet, name string) string {
    user := h.User().Search(rs.Env(), q.User().Name().Equals(name)).Limit(1)
    user.Sanitize()     
    return user.Email() 
}

func init() {
    h.User().NewMethod("GetEmail", user_GetEmail)
}
```
```go
package B

func init() {
    h.User().Methods().GetEmail().Extend(
    	func(rs m.UserSet, name string) string {
    	    res := rs.Super().GetEmail(name)
    	    return fmt.Sprintf("<%s>", res)
    	})
}
```
```go
// Anywhere else
email := h.User().NewSet(env).GetEmail("John")
fmt.Println(email)
// output: <john@example.com>
```

And it works also with views:
```xml
<!-- Package A -->
<view id="base_view_users_tree" model="User">
    <tree string="Users">
        <field name="Name"/>
        <field name="Login"/>
        <field name="Lang"/>
        <field name="LoginDate"/>
    </tree>
</view>
```
```xml
<!-- Package B -->
<view inherit_id="base_view_users_tree">
    <field name="Login" position="after">
        <field name="IsStaff"/>
    </field>
</view>
```
And the rendered view will be :
```xml
<tree string="Users">
    <field name="Name"/>
    <field name="Login"/>
    <field name="IsStaff"/>
    <field name="Lang"/>
    <field name="LoginDate"/>
</tree>
```

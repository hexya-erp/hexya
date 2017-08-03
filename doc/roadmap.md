Hexya Roadmap
===========

ORM
---
- [X] Rewrite ORM from scratch
- [X] Add methods to ORM models
- [X] Cache to RecordSets
- [X] Support for schema modification (ALTER TABLE)
- [X] Type safe API
- [X] Models inheritance
    - [X] 'Extention' model inheritance (ExtendModel)
    - [X] 'Embedding' model inheritance (embed)
    - [X] 'Mix In' model inheritance (MixInModel)
- [X] Relation fields
    - [X] Many2One relations
    - [X] One2One relations
    - [X] One2Many relations
    - [X] Rev2One relations
    - [X] Many2many relations
- [X] Computed fields to ORM:
    - [X] Fields computed by ERP after retrieval of computation vars
    - [X] Fields computed by ERP and stored in DB column
- [X] Related fields
    - [X] ReadOnly related fields
    - [X] ReadWrite related fields
    - [X] Searchable related fields
- [X] Security
    - [X] Permissions to execute methods
    - [X] CRUD permissions to fields
    - [X] Record rules on models
- [X] Handle serialization transaction isolation with appropriate retries
- [X] Improved model creation/extension API
- [X] Add more type safety in the condition builder
- [X] Support for SQL views and materialized views
- [X] Database foreign keys to related fields
- [X] Implement "group by" queries
- [X] Implement 'child_of' domain operator
- [X] Implement model constraints
- [X] Implement search restrictions for relation fields
- [ ] i18n and l10n support to ORM models
- [ ] Implement sending warning and domain with onchange
- [ ] Pagination API for RecordSets

Views
-----
- [X] Basic views
- [X] Inherited views
- [X] Manage embedded views

Controllers
-----------
- [X] Inheritable controllers

Server
------
- [X] Create controllers for using Odoo web client with Hexya Server
- [X] Recover from orm methods' panics
- [X] Unified logging system
- [X] Automate routing and include for `static` dir in modules
- [X] Improve hexya CLI with a cobra commander
- [ ] Implement hexya REPL console
- [ ] Redis cache for multi-server session store

Client
------
- [X] Adapt Odoo web client to be used in Hexya (V8)
- [X] Adapt Odoo web client to be used in Hexya (V9)
- [X] Rebrand web client

Modules
-------
- [X] Make module registering create necessary symlinks
- [X] Business logic testing framework
- [ ] Internal resource XML data files
    - [X] New schema for internal resources XML
    - [ ] Make hexya-generate create XSD for XML autocompletion
- [X] Add support for CSV data files
- [ ] Interface for report engines such as Jasper Reports

Documentation
-------------
- [X] Installation guide
- [X] Models API reference
- [X] Security reference
- [ ] Views reference
- [ ] Actions and menus reference
- [X] Module development tutorial
- [ ] Low level models API reference
- [ ] Testing Hexya
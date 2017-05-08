YEP Roadmap
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
    - [X] CRUD permissions to models and fields
    - [X] Record rules on models
- [X] Handle serialization transaction isolation with appropriate retries
- [X] Improved model creation/extension API
- [X] Add more type safety in the condition builder
- [X] Support for SQL views and materialized views
- [X] Database foreign keys to related fields
- [X] Implement "group by" queries
- [ ] Implement efficient 'child_of' domain operator
- [ ] Pagination API for RecordSets
- [ ] i18n and l10n support to ORM models

Views
-----
- [X] Basic views
- [X] Inherited views

Controllers
-----------
- [X] Inheritable controllers

Server
------
- [X] Create controllers for using Odoo web client with YEP Server
- [X] Recover from orm methods' panics
- [X] Unified logging system
- [X] Automate routing and include for `static` dir in modules
- [X] Improve yep CLI with a cobra commander

Client
------
- [X] Adapt Odoo web client to be used in YEP (V8)
- [X] Adapt Odoo web client to be used in YEP (V9)
- [ ] Rebrand web client

Modules
-------
- [X] Make module registering create necessary symlinks
- [X] Business logic testing framework
- [ ] Internal resource XML data files
    - [X] New schema for internal resources XML
    - [ ] Make yep-generate create XSD for XML autocompletion
- [X] Add support for CSV data files

Documentation
-------------
- [X] Installation guide
- [X] Models API reference
- [X] Security reference
- [ ] Views reference
- [ ] Actions and menus reference
- [ ] Low level models API reference
- [ ] Module development tutorial
- [ ] Testing YEP
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
- [ ] Related fields
    - [X] ReadOnly related fields
    - [X] ReadWrite related fields
    - [X] Searchable related fields
    - [ ] Related stored fields
- [ ] Security
    - [ ] CRUD permissions to models and fields
    - [ ] Record rules on models
- [ ] Handle serialization transaction isolation with appropriate retries
- [ ] Support for SQL views and materialized views
- [ ] Implement "group by" queries
- [ ] Implement efficient 'child_of' domain operator
- [ ] Database foreign keys to related fields
- [ ] i18n and l10n support to ORM models

Views
-----
- [X] Basic views
- [ ] Inherited views

Controllers
-----------
- [ ] Inheritable controllers

Server
------
- [X] Create controllers for using Odoo web client with YEP Server
- [X] Recover from orm methods' panics
- [X] Unified logging system
- [ ] Automate routing and include for `static` dir in modules

Client
------
- [X] Adapt Odoo web client to be used in YEP (V8)
- [ ] Adapt Odoo web client to be used in YEP (V9)

Modules
-------
- [X] Make module registering create necessary symlinks
- [ ] Internal resource XML data files
    - [X] Odoo like schema
    - [ ] New schema for internal resources XML
    - [ ] Make yep-generate create XSD for XML autocompletion
- [ ] Add support for data & demo XML files
- [ ] Add support for CSV data files
- [ ] Business logic testing framework

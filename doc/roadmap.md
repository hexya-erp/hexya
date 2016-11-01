YEP Roadmap
===========

ORM
---
- [X] Rewrite ORM from scratch
- [X] Add methods to ORM models
- [X] Relation fields
    - [X] Many2One relations
    - [X] One2One relations
    - [X] One2Many relations
    - [X] Rev2One relations
    - [X] Many2many relations
- [ ] Related fields
    - [X] ReadOnly related fields
    - [ ] ReadWrite related fields
    - [ ] Searchable related fields
- [ ] Models inheritance
    - [X] 'Extention' model inheritance (ExtendModel)
    - [X] 'Embedding' model inheritance (embed)
    - [ ] 'Mix In' model inheritance (MixInModel)
- [ ] Computed fields to ORM:
    - [X] Fields computed by ERP after retrieval of computation vars
    - [X] Fields computed by ERP and stored in DB column
    - [ ] Fields computed by DB by SQL function when reading DB
- [ ] Security
    - [ ] CRUD permissions to models and fields
    - [ ] Record rules on models
- [ ] i18n and l10n support to ORM models
- [ ] Database foreign keys to related fields
- [ ] Support for SQL views and materialized views
- [X] Cache to RecordSets
- [X] Support for schema modification (ALTER TABLE)
- [ ] Implement "group by" queries
- [X] Type safe API

Views
-----
- [ ] Inherited views

Controllers
-----------
- [ ] Inheritable controllers

Server
------
- [X] Create controllers for using Odoo web client with YEP Server
- [ ] Automate routing and include for `static` dir in modules
- [X] Recover from orm methods' panics
- [X] Unified logging system

Client
------
- [X] Adapt Odoo web client to be used in YEP (V8)
- [ ] Adapt Odoo web client to be used in YEP (V9)

Modules
-------
- [X] Make module registering create necessary symlinks
- [X] Add support for internal resources XML data files
- [ ] Add support for data & demo XML files
- [ ] Add support for CSV data files
- [ ] Business logic testing framework

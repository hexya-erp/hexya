YEP Roadmap
===========

ORM
---
- [X] Rewrite ORM from scratch
- [X] Add methods to ORM models
- [X] Many2One relations
- [X] One2One relations
- [ ] One2Many relations
- [ ] Rev2One relations
- [ ] Many2many relations
- [X] ReadOnly related fields
- [ ] ReadWrite related fields
- [ ] Searchable related fields
- [X] 'Inherits' model inheritance (Odoo-like)
- [ ] Computed fields to ORM:
    - [X] Fields computed by ERP after retrieval of computation vars
    - [X] Fields computed by ERP and stored in DB column
    - [ ] Fields computed by DB by SQL function when reading DB
- [ ] CRUD permissions to models and fields
- [ ] i18n and l10n support to ORM models
- [ ] Database foreign keys to related fields
- [ ] Cache to RecordSets
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

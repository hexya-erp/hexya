YEP Roadmap
===========

ORM
---
- [X] Rewrite ORM from scratch
- [X] Add methods to ORM models
- [X] Manage Many2One relations
- [X] Manage One2One relations
- [ ] Manage One2Many relations
- [ ] Manage Rev2One relations
- [ ] Manage Many2many relations
- [ ] Related fields
- [ ] 'Inherits' model inheritance (Odoo-like)
- [ ] Add computed fields to ORM:
    - [X] Fields computed by ERP after retrieval of computation vars
    - [X] Fields computed by ERP and stored in DB column
    - [ ] Fields computed by DB by SQL function when reading DB
- [ ] Add CRUD permissions to models and fields
- [ ] Add i18n and l10n support to ORM models
- [ ] Add database foreign keys to related fields
- [ ] Add cache to RecordSets
- [X] Add support for schema modification (ALTER TABLE)

Views
-----
- [ ] Add support for inherited views

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

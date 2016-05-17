YEP Roadmap
===========

ORM
---
- [X] Modify Beego ORM to enable layered models
- [X] Add meta support to ORM models/fields
- [X] Add high level RecordSet API
- [X] Add methods to ORM models
- [ ] Add computed fields to ORM:
    - [X] Fields computed by ERP after retrieval of computation vars
    - [X] Fields computed by ERP and stored in DB column
    - [ ] Fields computed by DB by SQL function when reading DB
- [ ] Add CRUD permissions to models and fields
- [ ] Add i18n and l10n support to ORM models
- [ ] Add database foreign keys to related fields
- [ ] Add cache to RecordSets
- [ ] Add support for schema modification (ALTER TABLE)

Server
------
- [ ] Create controllers for using Odoo web client with YEP Server
- [ ] Automate routing and include for `static` dir in modules
- [ ] Recover from orm methods' panics
- [ ] Unified logging system
- [ ] Create a JSON-RPC server with same protocol as Odoo.

Client
------
- [ ] Adapt Odoo web client to be used in YEP

Modules
-------
- [X] Make module registering create necessary symlinks
- [X] Add support for internal resources data files
- [ ] Add support for data & demo XML files
- [ ] Add support for CSV data files

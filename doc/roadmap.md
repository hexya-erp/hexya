YEP Roadmap
===========

ORM
---
- [X] Modify Beego ORM to enable layered models
- [ ] Add meta support to ORM models/fields
- [ ] Add methods to ORM models
- [ ] Add computed fields to ORM:
    - [ ] Fields computed by ERP after retrieval of computation vars
    - [ ] Fields computed by ERP and stored in DB column
    - [ ] Fields computed by DB by SQL function when reading DB
- [ ] Add CRUD permissions to models and fields
- [ ] Add i18n and l10n support to ORM models

Server
------
- [ ] Create a JSON-RPC server with same protocol as Odoo.
- [ ] Adapt to JSON-RPC over Websocket to speed up client-server
communications

Client
------
- [ ] Adapt Odoo web client to be used in YEP

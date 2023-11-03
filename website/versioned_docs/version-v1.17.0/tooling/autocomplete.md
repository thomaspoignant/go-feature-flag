---
sidebar_position: 20
title: Auto-complete
description: Flag configuration auto-complete
---

# Flag configuration auto-complete

GO Feature Flag offer a way to have auto-completion when you create a flag file.

To achieve this we publish a `jsonschema` on [schemastore](https://www.schemastore.org). The store is used by all major
IDE you can use such as `vscode`, `intelliJ`, ...

To enable it, you just have to use the extension `.goff.yaml` for your configuration file, and it will be automatically
available for you _(example: `flag.goff.yaml`)_.

---
sidebar_position: 30
title: 🪄 Auto-complete
description: Flag configuration auto-complete
---

# 🪄 Auto-complete

GO Feature Flag offers a way to have auto-completion while creating a flag file.

To achieve this we publish a `jsonschema` on [schemastore](https://www.schemastore.org). The store is used by all major IDEs such as `vscode`, `intelliJ`, and others.

To enable it, you just have to use the extension `.goff.yaml` for your configuration file, and it will be automatically available for you _(example: `flag.goff.yaml`)_.

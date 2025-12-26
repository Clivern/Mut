# API translations (gettext)

API error and success messages are translated using **gettext**. This directory is embedded at build time; the app loads `.po` files at runtime.

| Format | Role |
|--------|------|
| **.po** | Source and runtime file: edit these to change or add languages (e.g. `ar.po`). |

**Message format:** Each entry has a `msgid` (the key used in code, e.g. `not_authenticated`) and a `msgstr` (the text shown to the user). Keys match the arguments to `locale.TR(r, "key")` in the API.

**Adding a language:** Duplicate `en.po` as e.g. `ar.po`, translate the `msgstr` values, leave `msgid` as-is, and commit the new `.po` file here.

# Authoring JSON-LD

This template injects one `@graph` JSON-LD block into each page `<head>`.

Japanese is the canonical guide: see [docs/jsonld_ja.md](./jsonld_ja.md). This page is a short English summary.

## Where to write what

| Want | Write in |
| --- | --- |
| `WebSite`, `SoftwareSourceCode` (site-wide) | `site.meta.yaml` |
| `Organization` (site-wide) | `site.meta.yaml` → `organization` |
| Non-standard entities (site-wide) | `site.meta.yaml` → `jsonLdExtra` |
| `WebPage` basics | Markdown `title` / `description` |
| Role entity (`TechArticle`, …) | `schemaRole` or `jsonLd.entities` |
| Extra `WebPage` properties | `jsonLd.webPage` |
| Non-standard entities (this page) | `jsonLd.extra` |
| FAQ Q/A | Body `::faq-list` / `::faq-item` (not frontmatter) |

Two use cases:

- **(i) Add properties to standard entities** → `jsonLd.webPage` / `jsonLd.entities`
- **(ii) Add entities outside the standard set** → `jsonLdExtra` (site) / `jsonLd.extra` (page)

## Short form (`schemaRole`)

```yaml
---
title: Getting started
description: How to use this documentation site scaffold
schemaRole: TechArticle
---
```

Emits `WebPage` + `WebSite` + `SoftwareSourceCode` + `TechArticle`.

Allowed `schemaRole` values: `TechArticle` / `HowTo` / `FAQPage`.

## Full form (`jsonLd`)

Use `jsonLd` to add properties or multiple entities. See `content/*/api.md` for a worked example, and [docs/jsonld_ja.md](./jsonld_ja.md) for the full rules (merge, known vs unknown types, Organization, escape hatches, FAQ MDC collection).

## Samples

Under `content/{ja,en}/`:

- `index.md` — no `schemaRole` (`WebPage` only)
- `overview.md` / `install.md` / `getting-started.md` — `schemaRole: TechArticle`
- `api.md` — `jsonLd` full form (property merge + multiple entities)
- `tutorial.md` — `schemaRole: HowTo`
- `faq.md` — `schemaRole: FAQPage` + MDC

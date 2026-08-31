export type JsonLdObject = Record<string, unknown>;

/** One entity declared in frontmatter: `type` plus schema.org properties. */
export type JsonLdEntityInput = { type: string } & Record<string, unknown>;

/** The `jsonLd` frontmatter block. */
export type PageJsonLdInput = {
  webPage?: Record<string, unknown>;
  entities?: JsonLdEntityInput[];
  /** Escape hatch: entities emitted verbatim, with no defaults applied. */
  extra?: Record<string, unknown>[];
};

export type EntityBuildContext = {
  /** Resolved WebPage @id; role entities hang off it via isPartOf. */
  pageUrl: string;
  siteUrl: string;
  title: string;
  description?: string;
  inLanguage?: string;
  faqMainEntity?: JsonLdObject[];
  /** Set when site.meta.yaml declares an organization; used for publisher. */
  organizationId?: string;
};

const KNOWN_FRAGMENTS: Record<string, string> = {
  TechArticle: "article",
  HowTo: "howto",
  FAQPage: "faq",
};

function fragmentFor(type: string): string {
  return (
    KNOWN_FRAGMENTS[type] ||
    type
      .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "")
  );
}

/**
 * YAML parses unquoted dates into Date objects, which JSON.stringify would emit
 * with a time component. Collapse midnight UTC back to a date-only literal so
 * `datePublished: 2026-01-01` survives as written.
 */
export function normalizeJsonLdValue(value: unknown): unknown {
  if (value instanceof Date) {
    const iso = value.toISOString();
    return iso.endsWith("T00:00:00.000Z") ? iso.slice(0, 10) : iso;
  }
  if (Array.isArray(value)) {
    return value.map(normalizeJsonLdValue);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as JsonLdObject).map(([key, entry]) => [
        key,
        normalizeJsonLdValue(entry),
      ]),
    );
  }
  return value;
}

function baseEntity(type: string, ctx: EntityBuildContext): JsonLdObject {
  if (type === "TechArticle" || type === "HowTo") {
    const entity: JsonLdObject = {
      "@type": type,
      "@id": `${ctx.pageUrl}#${fragmentFor(type)}`,
      ...(type === "TechArticle"
        ? { headline: ctx.title }
        : { name: ctx.title }),
      isPartOf: { "@id": ctx.pageUrl },
      about: { "@id": `${ctx.siteUrl}/#software` },
    };
    if (ctx.organizationId) {
      entity.publisher = { "@id": ctx.organizationId };
    }
    if (ctx.description) {
      entity.description = ctx.description;
    }
    return entity;
  }

  if (type === "FAQPage") {
    const entity: JsonLdObject = {
      "@type": type,
      "@id": `${ctx.pageUrl}#${fragmentFor(type)}`,
      isPartOf: { "@id": ctx.pageUrl },
    };
    if (ctx.faqMainEntity?.length) {
      entity.mainEntity = ctx.faqMainEntity;
    }
    return entity;
  }

  // Unknown type: only scaffold what is valid for every schema.org type. Which
  // relationships apply (isPartOf, about, ...) depends on the type, so leave
  // those to the author.
  return {
    "@type": type,
    "@id": `${ctx.pageUrl}#${fragmentFor(type)}`,
  };
}

/**
 * Build one role entity: conventional defaults first, authored properties on
 * top. Returns null when the entity would carry no meaning.
 */
export function buildJsonLdEntity(
  input: JsonLdEntityInput,
  ctx: EntityBuildContext,
): JsonLdObject | null {
  const type = String(input.type || "").trim();
  if (!type) {
    return null;
  }

  const { type: _type, ...authored } = input;
  const entity: JsonLdObject = {
    ...baseEntity(type, ctx),
    ...(normalizeJsonLdValue(authored) as JsonLdObject),
  };

  if (type === "FAQPage" && !entity.mainEntity) {
    return null;
  }

  return entity;
}

export function buildWebPage(
  ctx: EntityBuildContext,
  authored?: Record<string, unknown>,
): JsonLdObject {
  const entity: JsonLdObject = {
    "@type": "WebPage",
    "@id": ctx.pageUrl,
    url: ctx.pageUrl,
    name: ctx.title,
    ...(ctx.inLanguage ? { inLanguage: ctx.inLanguage } : {}),
    isPartOf: { "@id": `${ctx.siteUrl}/#website` },
    about: { "@id": `${ctx.siteUrl}/#software` },
  };
  if (ctx.organizationId) {
    entity.publisher = { "@id": ctx.organizationId };
  }
  if (ctx.description) {
    entity.description = ctx.description;
  }
  return {
    ...entity,
    ...(normalizeJsonLdValue(authored || {}) as JsonLdObject),
  };
}

/**
 * Escape hatch entries are emitted as authored. Anything that is not a plain
 * object is dropped with a warning rather than corrupting the graph.
 */
export function sanitizeExtraEntities(
  entries: unknown,
  source: string,
): JsonLdObject[] {
  if (!entries) {
    return [];
  }
  if (!Array.isArray(entries)) {
    console.warn(`[doc-site] ${source} must be an array; ignoring.`);
    return [];
  }

  const result: JsonLdObject[] = [];
  for (const entry of entries) {
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      console.warn(`[doc-site] ${source} entry is not an object; skipping.`);
      continue;
    }
    result.push(normalizeJsonLdValue(entry) as JsonLdObject);
  }
  return result;
}

/**
 * Site-wide Organization. Only emitted when site.meta.yaml declares one, so
 * sites that have not filled it in keep their previous graph.
 */
export function buildOrganization(
  authored: Record<string, unknown> | null | undefined,
  siteUrl: string,
  siteName: string,
): JsonLdObject | null {
  if (!authored || Object.keys(authored).length === 0) {
    return null;
  }
  return {
    "@type": "Organization",
    "@id": `${siteUrl}/#organization`,
    name: siteName,
    url: `${siteUrl}/`,
    ...(normalizeJsonLdValue(authored) as JsonLdObject),
  };
}

function declaredEntities(jsonLd?: PageJsonLdInput): JsonLdEntityInput[] {
  return (jsonLd?.entities || []).filter((entry) => entry && entry.type);
}

/**
 * `jsonLd.entities` is the full form; `schemaRole` stays supported as the
 * single-entity shorthand.
 */
export function resolveEntityInputs(
  schemaRole: string | undefined,
  jsonLd?: PageJsonLdInput,
): JsonLdEntityInput[] {
  const declared = declaredEntities(jsonLd);
  if (declared.length) {
    if (schemaRole) {
      console.warn(
        `[doc-site] schemaRole "${schemaRole}" is ignored because jsonLd.entities is set.`,
      );
    }
    return declared;
  }
  return schemaRole ? [{ type: schemaRole }] : [];
}

/** Same resolution as above without the conflict warning. */
export function includesEntityType(
  type: string,
  schemaRole: string | undefined,
  jsonLd?: PageJsonLdInput,
): boolean {
  const declared = declaredEntities(jsonLd);
  if (declared.length) {
    return declared.some((entry) => entry.type === type);
  }
  return schemaRole === type;
}

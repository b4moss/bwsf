import { defineContentConfig, defineCollection, z } from "@nuxt/content";

/**
 * One JSON-LD entity. `type` is the only reserved key; everything else is
 * passed through as schema.org properties and merged over the conventional
 * defaults built by `useJsonLd()`.
 */
const jsonLdEntitySchema = z
  .object({ type: z.string() })
  .passthrough();

const pageSchema = z.object({
  /** Shorthand for a single entity. Equivalent to jsonLd.entities: [{ type }]. */
  schemaRole: z.enum(["TechArticle", "HowTo", "FAQPage"]).optional(),
  jsonLd: z
    .object({
      webPage: z.record(z.unknown()).optional(),
      entities: z.array(jsonLdEntitySchema).optional(),
      /** Emitted verbatim; no defaults are applied. */
      extra: z.array(z.record(z.unknown())).optional(),
    })
    .optional(),
});

export default defineContentConfig({
  collections: {
    content_ja: defineCollection({
      type: "page",
      source: {
        include: "ja/**",
        prefix: "",
      },
      schema: pageSchema,
    }),
    content_en: defineCollection({
      type: "page",
      source: {
        include: "en/**",
        prefix: "",
      },
      schema: pageSchema,
    }),
  },
});

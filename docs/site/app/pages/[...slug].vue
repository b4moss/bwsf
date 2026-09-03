<script setup lang="ts">
import { withLeadingSlash } from "ufo";
import type { Collections } from "@nuxt/content";
import type { SchemaRole } from "~/composables/useJsonLd";
import { extractFaqFromBody, type FaqQa } from "~/utils/extractFaq";
import {
  includesEntityType,
  type PageJsonLdInput,
} from "~/utils/jsonLdEntities";

const route = useRoute();
const { locale } = useI18n();
const config = useRuntimeConfig();

/**
 * Content paths and useAsyncData keys must not depend on whether the static
 * host served `/ja/faq` or `/ja/faq/`. Trailing-slash drift after hydration
 * remounts the page with a new key, misses the payload cache, and throws 404 —
 * which removes the FAQ DOM (and looks like "tap does nothing" on mobile).
 */
const slug = computed(() => {
  const raw = route.params.slug;
  if (!raw || (Array.isArray(raw) && raw.length === 0)) {
    return "/";
  }
  const joined = Array.isArray(raw) ? raw.join("/") : String(raw);
  const withSlash = withLeadingSlash(joined);
  return withSlash === "/" ? "/" : withSlash.replace(/\/+$/, "");
});

/** Ignore browser / tooling asset probes (e.g. manifest.webmanifest). */
const isAssetPath = computed(() =>
  /\.[a-z0-9]{2,8}$/i.test(slug.value),
);

if (isAssetPath.value) {
  throw createError({
    statusCode: 404,
    statusMessage: "Not Found",
    fatal: false,
  });
}

const { data: page } = await useAsyncData(
  () => `content-${locale.value}-${slug.value}`,
  async () => {
    const collection = (`content_${locale.value}`) as keyof Collections;
    return queryCollection(collection).path(slug.value).first();
  },
  { watch: [locale, slug] },
);

if (!page.value) {
  throw createError({
    statusCode: 404,
    statusMessage: "Page not found",
    fatal: false,
  });
}

const pageTitle = computed(
  () => page.value?.title || String(config.public.siteName || "Doc Site"),
);

useSeoMeta({
  title: () => pageTitle.value,
  description: () => page.value?.description || undefined,
});

const siteUrl = computed(() =>
  String(config.public.siteUrl || "https://example.com").replace(/\/$/, ""),
);

const pageUrl = computed(() => {
  const path =
    slug.value === "/" ? `/${locale.value}` : `/${locale.value}${slug.value}`;
  return `${siteUrl.value}${path}`;
});

const schemaRole = computed(() => {
  const role = (page.value as { schemaRole?: SchemaRole } | null)?.schemaRole;
  return role;
});

const jsonLd = computed(
  () => (page.value as { jsonLd?: PageJsonLdInput } | null)?.jsonLd,
);

const hasFaqEntity = computed(() =>
  includesEntityType("FAQPage", schemaRole.value, jsonLd.value),
);

// Read from the content AST, not from rendered components, so the JSON-LD is
// per-page and settled before render.
const faqItems = computed<FaqQa[]>(() => {
  if (!hasFaqEntity.value || !page.value?.body) {
    return [];
  }
  return extractFaqFromBody(page.value.body);
});
</script>

<template>
  <div>
    <article class="prose">
      <ContentRenderer v-if="page" :value="page" />
    </article>
    <DocsJsonLd
      :page-url="pageUrl"
      :title="pageTitle"
      :description="page?.description || undefined"
      :schema-role="schemaRole"
      :json-ld="jsonLd"
      :faq-items="faqItems"
    />
    <DocsPager />
  </div>
</template>

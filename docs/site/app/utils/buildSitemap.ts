import { docsNavItems } from "~/config/docsNav";

export const sitemapLocales = ["ja", "en"] as const;

export type SitemapLocale = (typeof sitemapLocales)[number];

function escapeXml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

/**
 * Build a urlset sitemap from docsNav + locales (hreflang alternates included).
 */
export function buildSitemapXml(
  siteUrl: string,
  defaultLocale: SitemapLocale = "ja",
): string {
  const base = siteUrl.replace(/\/$/, "") || "https://example.com";

  const urls = docsNavItems.map((item) => {
    const path = item.path === "/" ? "" : item.path;
    const locByLocale = Object.fromEntries(
      sitemapLocales.map((locale) => [locale, `${base}/${locale}${path}`]),
    ) as Record<SitemapLocale, string>;

    const loc = locByLocale[defaultLocale];
    const alternates = sitemapLocales
      .map(
        (locale) =>
          `    <xhtml:link rel="alternate" hreflang="${locale}" href="${escapeXml(locByLocale[locale])}" />`,
      )
      .join("\n");
    const xDefault = `    <xhtml:link rel="alternate" hreflang="x-default" href="${escapeXml(locByLocale[defaultLocale])}" />`;

    return `  <url>
    <loc>${escapeXml(loc)}</loc>
${alternates}
${xDefault}
  </url>`;
  });

  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls.join("\n")}
</urlset>
`;
}

import { buildSitemapXml } from "~/utils/buildSitemap";

export default defineEventHandler((event) => {
  const { siteUrl } = useRuntimeConfig().public;
  const xml = buildSitemapXml(String(siteUrl || "https://example.com"));
  setHeader(event, "Content-Type", "application/xml; charset=utf-8");
  return xml;
});

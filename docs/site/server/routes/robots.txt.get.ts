export default defineEventHandler((event) => {
  const { siteUrl } = useRuntimeConfig().public;
  const base = String(siteUrl || "https://example.com").replace(/\/$/, "");
  setHeader(event, "Content-Type", "text/plain; charset=utf-8");
  return `User-agent: *\nAllow: /\n\nSitemap: ${base}/sitemap.xml\n`;
});

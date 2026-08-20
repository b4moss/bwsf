const LOCALES = new Set(["ja", "en"]);
const COOKIE_KEY = "i18n_redirected";
const FALLBACK = "ja";

function pickFromAcceptLanguage(header: string | undefined): string {
  if (!header) return FALLBACK;
  const parts = header.split(",").map((part) => {
    const [tag, ...params] = part.trim().split(";");
    const q = params.find((p) => p.trim().startsWith("q="));
    const quality = q ? Number(q.split("=")[1]) || 0 : 1;
    return { tag: tag.toLowerCase(), quality };
  });
  parts.sort((a, b) => b.quality - a.quality);

  for (const { tag } of parts) {
    if (tag.startsWith("ja")) return "ja";
    if (tag.startsWith("en")) return "en";
  }
  return FALLBACK;
}

export default defineEventHandler((event) => {
  const { pathname } = getRequestURL(event);
  if (pathname !== "/" && pathname !== "") return;

  const cookie = getCookie(event, COOKIE_KEY);
  const locale =
    cookie && LOCALES.has(cookie)
      ? cookie
      : pickFromAcceptLanguage(getHeader(event, "accept-language"));

  return sendRedirect(event, `/${locale}/`, 302);
});

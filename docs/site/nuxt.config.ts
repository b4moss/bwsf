import { copyFileSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

/** Latest product version from app/src/cmd/version.go (single source of truth). */
function readAppVersion(): string {
  const repoRoot = join(fileURLToPath(new URL(".", import.meta.url)), "../..");
  const versionGo = readFileSync(
    join(repoRoot, "app/src/cmd/version.go"),
    "utf8",
  );
  const match = versionGo.match(/const Version = "([^"]+)"/);
  if (!match) {
    throw new Error("Could not parse Version from app/src/cmd/version.go");
  }
  return `v${match[1]}`;
}

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    "@nuxt/content",
    "@nuxtjs/i18n",
    "@nuxtjs/color-mode",
    "@nuxt/scripts",
  ],
  devtools: { enabled: true },
  compatibilityDate: "2024-04-03",
  css: ["~/assets/css/main.css"],
  runtimeConfig: {
    public: {
      siteName: "bwsf",
      siteVersion: readAppVersion(),
      githubUrl: "https://github.com/b4moss/bwsf",
      footerText: "MIT License · 2026 Bicycle for Mind LLC.",
    },
  },
  // GTM: set NUXT_PUBLIC_SCRIPTS_GOOGLE_TAG_MANAGER_ID=GTM-XXXXXXX (build-time for SSG).
  // Empty / unset → tagging stays disabled (see plugins/google-tag-manager.client.ts).
  // Analytics for this site uses GA via CookieConsent (G-CNCM19EPXM), not GTM.
  scripts: {
    registry: {
      googleTagManager: {
        bundle: false,
      },
    },
  },
  colorMode: {
    preference: "system",
    fallback: "light",
    classSuffix: "",
  },
  content: {
    // Avoid better-sqlite3 native bindings on Netlify CI (Node 22+)
    experimental: { sqliteConnector: "native" },
    build: {
      markdown: {
        highlight: {
          theme: {
            default: "github-light",
            dark: "github-dark",
          },
        },
      },
    },
  },
  app: {
    head: {
      meta: [{ name: "theme-color", content: "#175ddc" }],
      link: [
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap",
        },
      ],
    },
  },
  i18n: {
    locales: [
      { code: "ja", name: "日本語", language: "ja-JP", file: "ja.ts" },
      { code: "en", name: "English", language: "en-US", file: "en.ts" },
    ],
    defaultLocale: "ja",
    strategy: "prefix",
    lazy: true,
    langDir: "locales",
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: "i18n_redirected",
      redirectOn: "root",
      fallbackLocale: "ja",
    },
    bundle: {
      optimizeTranslationDirective: false,
    },
  },
  // public/index.html would shadow `/` in `nuxt dev` and block Nitro middleware.
  // Copy the static locale redirect page into the generate output instead.
  hooks: {
    "nitro:build:public-assets"(nitro) {
      copyFileSync(
        join(nitro.options.rootDir, "locale-root.html"),
        join(nitro.options.output.publicDir, "index.html"),
      );
    },
  },
  nitro: {
    preset: "static",
    prerender: {
      crawlLinks: true,
      routes: [
        "/ja",
        "/en",
        "/ja/guide/getting-started",
        "/en/guide/getting-started",
        "/ja/guide/features",
        "/en/guide/features",
        "/ja/guide/installation",
        "/en/guide/installation",
        "/ja/guide/commands",
        "/en/guide/commands",
        "/ja/guide/philosophy",
        "/en/guide/philosophy",
        "/ja/guide/upgrade",
        "/en/guide/upgrade",
        "/ja/guide/uninstall",
        "/en/guide/uninstall",
        "/ja/guide/dev-loadmap",
        "/en/guide/dev-loadmap",
        "/ja/guide/faq",
        "/en/guide/faq",
        "/ja/guide/license",
        "/en/guide/license",
        "/ja/guide/license-faq",
        "/en/guide/license-faq",
        "/ja/cookie-policy",
        "/en/cookie-policy",
      ],
    },
  },
});

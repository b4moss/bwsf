export type DocsNavItem = {
  key: string;
  path: string;
  /** i18n key under `nav.*` (e.g. `home` → `nav.home`). */
  labelKey: string;
  /** When set, item is a child of this nav key (shown indented in sidebar). */
  parent?: string;
};

/**
 * Edit this list to shape the docs sidebar / pager.
 * Labels come from `i18n/locales/{ja,en}.ts` → `nav.<labelKey>`.
 */
export const docsNavItems: DocsNavItem[] = [
  { key: "home", path: "/", labelKey: "home" },
  { key: "gettingStarted", path: "/guide/getting-started", labelKey: "gettingStarted" },
  { key: "features", path: "/guide/features", labelKey: "features" },
  { key: "installation", path: "/guide/installation", labelKey: "installation" },
  { key: "commands", path: "/guide/commands", labelKey: "commands" },
  { key: "philosophy", path: "/guide/philosophy", labelKey: "philosophy" },
  { key: "upgrade", path: "/guide/upgrade", labelKey: "upgrade", parent: "other" },
  { key: "uninstall", path: "/guide/uninstall", labelKey: "uninstall", parent: "other" },
  { key: "devLoadmap", path: "/guide/dev-loadmap", labelKey: "devLoadmap", parent: "other" },
  { key: "faq", path: "/guide/faq", labelKey: "faq" },
  { key: "license", path: "/guide/license", labelKey: "license" },
  { key: "licenseFaq", path: "/guide/license-faq", labelKey: "licenseFaq" },
  { key: "cookiePolicy", path: "/cookie-policy", labelKey: "cookiePolicy" },
];

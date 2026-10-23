import { docsNavItems } from "~/config/docsNav";

function normalizePath(path: string) {
  if (!path || path === "/") return "/";
  return path.replace(/\/+$/, "") || "/";
}

export function useDocsNav() {
  const { t } = useI18n();
  const localePath = useLocalePath();

  const items = computed(() =>
    docsNavItems.map((item) => {
      const external = Boolean(item.externalHref);
      return {
        ...item,
        label: t(`nav.${item.labelKey}`),
        to: external ? String(item.externalHref) : localePath(item.path || "/"),
        external,
      };
    }),
  );

  /** In-site pages only (pager / active-route). */
  const pageItems = computed(() => items.value.filter((item) => !item.external));

  return { items, pageItems };
}

export function useDocsPager() {
  const route = useRoute();
  const { pageItems } = useDocsNav();

  const index = computed(() => {
    const current = normalizePath(route.path);
    return pageItems.value.findIndex(
      (item) => normalizePath(String(item.to)) === current,
    );
  });

  const prev = computed(() => {
    const i = index.value;
    return i > 0 ? pageItems.value[i - 1] : null;
  });

  const next = computed(() => {
    const i = index.value;
    return i >= 0 && i < pageItems.value.length - 1 ? pageItems.value[i + 1] : null;
  });

  return { prev, next };
}

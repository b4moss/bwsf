import { docsNavItems } from "~/config/docsNav";

function normalizePath(path: string) {
  if (!path || path === "/") return "/";
  return path.replace(/\/+$/, "") || "/";
}

export function useDocsNav() {
  const { t } = useI18n();
  const localePath = useLocalePath();

  const items = computed(() =>
    docsNavItems.map((item) => ({
      ...item,
      label: t(`nav.${item.labelKey}`),
      to: localePath(item.path),
    })),
  );

  return { items };
}

export function useDocsPager() {
  const route = useRoute();
  const { items } = useDocsNav();

  const index = computed(() => {
    const current = normalizePath(route.path);
    return items.value.findIndex(
      (item) => normalizePath(String(item.to)) === current,
    );
  });

  const prev = computed(() => {
    const i = index.value;
    return i > 0 ? items.value[i - 1] : null;
  });

  const next = computed(() => {
    const i = index.value;
    return i >= 0 && i < items.value.length - 1 ? items.value[i + 1] : null;
  });

  return { prev, next };
}

<script setup lang="ts">
type LocaleOption = {
  code: string;
  name?: string;
};

const { t, locale, locales, setLocale } = useI18n();
const colorMode = useColorMode();

const languageOptions = computed(() => {
  const allowed = new Set(["ja", "en"]);
  const labels: Record<string, string> = {
    en: "English",
    ja: "日本語",
  };

  return (locales.value as LocaleOption[])
    .filter((item) => allowed.has(item.code))
    .map((item) => ({
      code: item.code,
      name: labels[item.code] ?? item.name ?? item.code,
    }));
});

const themeOptions = computed(() => [
  { value: "system", label: t("theme.system"), compact: "Sys" },
  { value: "light", label: t("theme.light"), compact: "Lt" },
  { value: "dark", label: t("theme.dark"), compact: "Dk" },
]);

const currentLanguage = computed(() => {
  const current = languageOptions.value.find((item) => item.code === locale.value);
  return current?.name ?? locale.value;
});

const currentTheme = computed(
  () =>
    themeOptions.value.find((item) => item.value === colorMode.preference) ??
    themeOptions.value[0],
);

const triggerText = computed(
  () => `${currentLanguage.value} · ${currentTheme.value.label}`,
);

const compactText = computed(() => {
  const lang =
    locale.value === "ja" ? "JA" : locale.value === "en" ? "EN" : locale.value.toUpperCase();
  return `${lang} · ${currentTheme.value.compact}`;
});

async function chooseLanguage(code: string, close: () => void) {
  if (code !== locale.value) {
    await setLocale(code);
  }
  close();
}

function chooseTheme(value: string, close: () => void) {
  colorMode.preference = value;
  close();
}
</script>

<template>
  <HeaderDropdown
    :label="t('nav.prefs')"
    :trigger-text="triggerText"
    :compact-text="compactText"
  >
    <template #default="{ close }">
      <div class="section" role="group" :aria-label="t('nav.language')">
        <p class="section-label">{{ t("nav.language") }}</p>
        <button
          v-for="item in languageOptions"
          :key="item.code"
          type="button"
          class="option"
          role="option"
          :aria-selected="item.code === locale"
          :data-active="item.code === locale ? 'true' : 'false'"
          @click="chooseLanguage(item.code, close)"
        >
          {{ item.name }}
        </button>
      </div>
      <div class="section" role="group" :aria-label="t('theme.label')">
        <p class="section-label">{{ t("theme.label") }}</p>
        <button
          v-for="item in themeOptions"
          :key="item.value"
          type="button"
          class="option"
          role="option"
          :aria-selected="item.value === colorMode.preference"
          :data-active="item.value === colorMode.preference ? 'true' : 'false'"
          @click="chooseTheme(item.value, close)"
        >
          {{ item.label }}
        </button>
      </div>
    </template>
  </HeaderDropdown>
</template>

<style scoped>
.section + .section {
  margin-top: 0.35rem;
  padding-top: 0.35rem;
  border-top: 1px solid var(--color-border);
}

.section-label {
  margin: 0;
  padding: 0.25rem 0.65rem 0.15rem;
  color: var(--color-muted);
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.option {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 2.1rem;
  padding: 0.4rem 0.65rem;
  border: 0;
  border-radius: 0.3rem;
  background: transparent;
  color: var(--color-ink);
  font: inherit;
  font-size: 0.875rem;
  text-align: left;
  cursor: pointer;
}

.option:hover {
  background: var(--color-accent-soft);
}

.option[data-active="true"] {
  color: var(--color-accent);
  font-weight: 600;
}

@media (max-width: 640px) {
  .option {
    min-height: 2.4rem;
    font-size: 0.95rem;
  }
}
</style>

<script setup lang="ts">
const { t } = useI18n();
const localePath = useLocalePath();
const { open, toggle } = useSidebar();
const config = useRuntimeConfig();

const siteName = computed(() => String(config.public.siteName || "Doc Site"));
const siteVersion = computed(() => String(config.public.siteVersion || ""));
const githubUrl = computed(() => String(config.public.githubUrl || ""));
</script>

<template>
  <header class="header">
    <div class="header-inner">
      <button
        type="button"
        class="menu-btn"
        :aria-expanded="open"
        :aria-label="t('nav.menu')"
        @click="toggle"
      >
        <span class="menu-icon" aria-hidden="true" />
      </button>
      <div class="masthead">
        <NuxtLink :to="localePath('/')" class="brand">
          {{ siteName }}
        </NuxtLink>
        <p
          v-if="siteVersion"
          class="versions"
          :aria-label="t('nav.versions')"
        >
          <span>{{ siteVersion }}</span>
        </p>
      </div>
      <div class="header-actions">
        <HeaderPrefsMenu />
        <a
          v-if="githubUrl"
          class="github-link"
          :href="githubUrl"
          target="_blank"
          rel="noopener noreferrer"
          :aria-label="t('nav.github')"
          :title="t('nav.github')"
        >
          <svg
            class="github-icon"
            viewBox="0 0 16 16"
            width="20"
            height="20"
            aria-hidden="true"
          >
            <path
              fill="currentColor"
              d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8"
            />
          </svg>
        </a>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 30;
  height: var(--header-height);
  backdrop-filter: blur(10px);
  background: color-mix(in srgb, var(--color-bg) 82%, var(--color-surface));
  border-bottom: 1px solid var(--color-border);
}

.header-inner {
  height: 100%;
  width: 100%;
  padding: 0 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.menu-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--color-border);
  border-radius: 0.35rem;
  background: var(--color-surface);
  cursor: pointer;
  padding: 0;
}

.menu-icon,
.menu-icon::before,
.menu-icon::after {
  display: block;
  width: 1rem;
  height: 2px;
  background: var(--color-ink);
  border-radius: 1px;
}

.menu-icon {
  position: relative;
}

.menu-icon::before,
.menu-icon::after {
  content: "";
  position: absolute;
  left: 0;
}

.menu-icon::before {
  top: -5px;
}

.menu-icon::after {
  top: 5px;
}

.masthead {
  display: flex;
  flex-direction: row;
  align-items: baseline;
  gap: 0.55rem;
  margin-right: auto;
  min-width: 0;
}

.brand {
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--color-ink);
  text-decoration: none;
  white-space: nowrap;
  line-height: 1.2;
}

.brand:hover {
  color: var(--color-accent);
}

.versions {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 0.7rem;
  font-weight: 400;
  letter-spacing: 0.01em;
  color: var(--color-muted);
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.github-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.35rem;
  color: var(--color-muted);
  text-decoration: none;
}

.github-link:hover {
  color: var(--color-ink);
  background: var(--color-accent-soft);
}

@media (min-width: 900px) {
  .menu-btn {
    display: none;
  }

  .header-inner {
    padding: 0 1.25rem;
  }
}

@media (max-width: 420px) {
  .versions {
    font-size: 0.6rem;
  }
}
</style>

<script setup lang="ts">
const { t } = useI18n();
const { prev, next } = useDocsPager();
</script>

<template>
  <nav v-if="prev || next" class="docs-pager" :aria-label="t('nav.pager')">
    <NuxtLink v-if="prev" :to="prev.to" class="pager-link pager-prev">
      <span class="pager-dir">{{ t("nav.prev") }}</span>
      <span class="pager-title">{{ prev.label }}</span>
    </NuxtLink>
    <span v-else class="pager-spacer" />
    <NuxtLink v-if="next" :to="next.to" class="pager-link pager-next">
      <span class="pager-dir">{{ t("nav.next") }}</span>
      <span class="pager-title">{{ next.label }}</span>
    </NuxtLink>
    <span v-else class="pager-spacer" />
  </nav>
</template>

<style scoped>
.docs-pager {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  width: min(100%, var(--max-width));
  margin-top: 3rem;
  margin-inline: auto;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border);
}

.pager-link {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  background: var(--color-surface);
  text-decoration: none;
  color: var(--color-ink);
  min-width: 0;
}

.pager-link:hover {
  border-color: var(--color-accent);
  background: var(--color-accent-soft);
}

.pager-dir {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-muted);
}

.pager-title {
  font-size: 0.95rem;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pager-next {
  text-align: right;
  grid-column: 2;
}

.pager-prev {
  grid-column: 1;
}

.pager-spacer {
  display: none;
}

@media (max-width: 520px) {
  .docs-pager {
    grid-template-columns: 1fr;
  }

  .pager-prev,
  .pager-next {
    grid-column: 1;
    text-align: left;
  }
}
</style>

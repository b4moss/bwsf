<script setup lang="ts">
const { items } = useDocsNav();
const { open, close } = useSidebar();
const route = useRoute();

watch(
  () => route.fullPath,
  () => {
    close();
  },
);

function isActive(item: { to: string }) {
  const current = route.path.replace(/\/+$/, "") || "/";
  const target = String(item.to).replace(/\/+$/, "") || "/";
  return current === target;
}
</script>

<template>
  <aside class="sidebar" :class="{ open }" aria-label="Docs">
    <nav class="sidebar-nav">
      <NuxtLink
        v-for="item in items"
        :key="item.key"
        :to="item.to"
        class="sidebar-link"
        :class="{
          'sidebar-link--child': Boolean(item.parent),
          'router-link-exact-active': isActive(item),
        }"
        @click="close"
      >
        {{ item.label }}
      </NuxtLink>
    </nav>
  </aside>
  <button
    v-if="open"
    type="button"
    class="sidebar-backdrop"
    aria-label="Close menu"
    @click="close"
  />
</template>

<style scoped>
.sidebar {
  position: fixed;
  top: var(--header-height);
  left: 0;
  bottom: 0;
  z-index: 20;
  width: var(--sidebar-width);
  padding: 1.25rem 1rem;
  background: color-mix(in srgb, var(--color-bg) 92%, var(--color-surface));
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
  transform: translateX(-100%);
  transition: transform 0.2s ease;
}

.sidebar.open {
  transform: translateX(0);
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.sidebar-link {
  display: block;
  padding: 0.45rem 0.7rem;
  border-radius: 0.35rem;
  color: var(--color-muted);
  text-decoration: none;
  font-size: 0.9375rem;
  font-weight: 500;
}

.sidebar-link:hover {
  color: var(--color-ink);
  background: var(--color-accent-soft);
}

.sidebar-link--child {
  padding-left: 1.35rem;
  font-size: 0.875rem;
  font-weight: 400;
}

.sidebar-link.router-link-exact-active {
  color: var(--color-accent);
  background: var(--color-accent-soft);
  font-weight: 600;
}

.sidebar-backdrop {
  position: fixed;
  inset: var(--header-height) 0 0;
  z-index: 15;
  border: none;
  padding: 0;
  background: color-mix(in srgb, var(--color-ink) 35%, transparent);
  cursor: pointer;
}

@media (min-width: 900px) {
  .sidebar {
    position: sticky;
    top: var(--header-height);
    height: calc(100vh - var(--header-height));
    transform: none;
    flex-shrink: 0;
  }

  .sidebar-backdrop {
    display: none;
  }
}
</style>

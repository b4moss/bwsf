<script setup lang="ts">
const ossLinks = [
  { name: "bwsf", href: "https://bwsf.oss.b4m.jp/" },
  { name: "jp-local-gov-id", href: "https://jplocalgov.oss.b4m.jp" },
  { name: "necromancer", href: "https://github.com/b4moss/" },
  { name: "crudian", href: "https://github.com/b4moss/crudian" },
  { name: "shardian", href: "https://github.com/b4moss/shardian" },
  { name: "janitor", href: "https://github.com/b4moss/janitor" },
] as const;
</script>

<template>
  <footer class="site-footer">
    <nav class="site-footer-oss" aria-label="OSS products">
      <p class="site-footer-oss-label">b4moss' friend:</p>
      <ul class="site-footer-oss-list">
        <li v-for="link in ossLinks" :key="link.name">
          <a :href="link.href" target="_blank" rel="noopener noreferrer">
            {{ link.name }}
          </a>
        </li>
      </ul>
    </nav>
    <p class="site-footer-text">
      MIT License · 2026
      <a href="https://b4m.co.jp/" target="_blank" rel="noopener noreferrer"
        >Bicycle for Mind LLC.</a
      >
    </p>
  </footer>
</template>

<style scoped>
.site-footer {
  border-top: 1px solid var(--color-border);
  padding: 1.25rem 1.5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.site-footer-oss {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: center;
  gap: 0.35rem 0.65rem;
}

.site-footer-oss-label {
  margin: 0;
  font-size: 0.75rem;
  color: var(--color-muted);
  opacity: 0.75;
}

.site-footer-oss-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.35rem 0.85rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.site-footer-oss-list a {
  font-size: 0.8125rem;
  font-family: var(--font-mono);
  color: var(--color-muted);
  text-decoration: none;
}

.site-footer-oss-list a:hover {
  color: var(--color-accent);
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.15em;
}

.site-footer-text {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-muted);
}
</style>

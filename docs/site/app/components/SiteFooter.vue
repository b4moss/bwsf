<script setup lang="ts">
/*
const config = useRuntimeConfig();
const footerText = computed(
  () =>
    String(config.public.footerText || "") ||
    `MIT License · 2026 <a href="https://b4m.co.jp/" target="_blank">Bicycle for Mind LLC.</a>`,
);
*/
</script>

<template>
  <footer class="site-footer">
    <p class="site-footer-text">MIT License · 2026 <a href="https://b4m.co.jp/" target="_blank">Bicycle for Mind LLC.</a></p>
  </footer>
</template>

<style scoped>
.site-footer {
  border-top: 1px solid var(--color-border);
  padding: 1.25rem 1.5rem;
  text-align: center;
}

.site-footer-text {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-muted);
}
</style>

<script setup lang="ts">
const config = useRuntimeConfig();

const siteName = computed(() => String(config.public.siteName || "bwsf"));

useHead({
  titleTemplate: (title) =>
    title ? `${title} :: ${siteName.value}` : siteName.value,
});

useGoogleAnalytics();
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <CookieConsent />
</template>

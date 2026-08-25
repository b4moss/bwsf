export default defineNuxtPlugin(() => {
  // Never load GTM from local `nuxt dev` / development builds.
  if (import.meta.dev) {
    return;
  }

  const id = useRuntimeConfig().public.scripts?.googleTagManager?.id;
  if (typeof id !== "string" || !id.startsWith("GTM-")) {
    return;
  }

  useScriptGoogleTagManager({
    id,
    scriptOptions: {
      trigger: "onNuxtReady",
      // Serve live gtm.js from Google (not a build-time bundle snapshot).
      bundle: false,
    },
  });
});

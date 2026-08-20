import { defineContentConfig, defineCollection } from "@nuxt/content";

export default defineContentConfig({
  collections: {
    content_ja: defineCollection({
      type: "page",
      source: {
        include: "ja/**",
        prefix: "",
      },
    }),
    content_en: defineCollection({
      type: "page",
      source: {
        include: "en/**",
        prefix: "",
      },
    }),
  },
});

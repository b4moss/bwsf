<script lang="ts">
import { defineComponent, toRef, type PropType } from "vue";
import type { SchemaRole } from "~/composables/useJsonLd";
import type { PageJsonLdInput } from "~/utils/jsonLdEntities";
import type { FaqQa } from "~/utils/extractFaq";

/**
 * Head-only helper: JSON-LD is injected via useHead.
 * Must render `null` (not an empty / comment-only template) so SSR and client
 * agree on a single comment anchor — empty templates are stripped in production
 * builds and cause hydration mismatches that wipe the page body.
 */
export default defineComponent({
  name: "DocsJsonLd",
  props: {
    pageUrl: { type: String, required: true },
    title: { type: String, required: true },
    description: { type: String, required: false },
    schemaRole: { type: String as PropType<SchemaRole>, required: false },
    jsonLd: { type: Object as PropType<PageJsonLdInput>, required: false },
    faqItems: { type: Array as PropType<FaqQa[]>, required: false },
  },
  setup(props) {
    useJsonLd({
      pageUrl: toRef(props, "pageUrl"),
      title: toRef(props, "title"),
      description: toRef(props, "description"),
      schemaRole: toRef(props, "schemaRole"),
      jsonLd: toRef(props, "jsonLd"),
      faqItems: toRef(props, "faqItems"),
    });
    return () => null;
  },
});
</script>

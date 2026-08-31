<script setup lang="ts">
import {
  faqListInjectionKey,
  type FaqListContext,
} from "~/utils/faqListContext";

defineProps<{
  question: string;
}>();

const id = useId();
const faqList = inject<FaqListContext | null>(faqListInjectionKey, null);

onMounted(() => {
  faqList?.registerPanel(id);
});

onBeforeUnmount(() => {
  faqList?.unregisterPanel(id);
});

const open = computed(() => faqList?.isOpen(id) ?? false);

function onToggle() {
  faqList?.toggle(id);
}
</script>

<template>
  <div class="faq-item" :data-open="open">
    <h3 class="faq-item__question">
      <button
        class="faq-item__trigger"
        type="button"
        :aria-expanded="open"
        :aria-controls="`${id}-panel`"
        :id="`${id}-trigger`"
        @click="onToggle"
      >
        <span>{{ question }}</span>
        <span class="faq-item__chevron" aria-hidden="true" />
      </button>
    </h3>
    <div
      class="faq-item__panel"
      :id="`${id}-panel`"
      role="region"
      :aria-labelledby="`${id}-trigger`"
      :aria-hidden="!open"
    >
      <div class="faq-item__panel-inner">
        <div class="faq-item__answer">
          <slot />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.faq-item {
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  background: var(--color-surface);
}

.faq-item__question {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.faq-item__trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin: 0;
  padding: 0.85rem 1rem;
  border: none;
  background: transparent;
  color: inherit;
  font: inherit;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}

.faq-item__trigger:hover {
  color: var(--color-accent);
}

.faq-item__chevron {
  position: relative;
  flex-shrink: 0;
  width: 0.7rem;
  height: 0.7rem;
  border-right: 1.5px solid var(--color-muted);
  border-bottom: 1.5px solid var(--color-muted);
  transform: rotate(45deg);
  transition: transform 0.28s ease, border-color 0.2s ease;
}

.faq-item__trigger:hover .faq-item__chevron {
  border-color: var(--color-accent);
}

.faq-item[data-open="true"] .faq-item__chevron {
  transform: rotate(225deg);
}

.faq-item__panel {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s ease;
}

.faq-item[data-open="true"] .faq-item__panel {
  grid-template-rows: 1fr;
}

.faq-item__panel-inner {
  overflow: hidden;
  min-height: 0;
}

.faq-item__answer {
  padding: 0 1rem 1rem;
  color: var(--color-ink);
  opacity: 0;
  transform: translateY(-0.2rem);
  transition:
    opacity 0.22s ease,
    transform 0.28s ease;
}

.faq-item[data-open="true"] .faq-item__answer {
  opacity: 1;
  transform: none;
}

.faq-item__answer :deep(> *:first-child) {
  margin-top: 0;
}

.faq-item__answer :deep(> *:last-child) {
  margin-bottom: 0;
}

@media (prefers-reduced-motion: reduce) {
  .faq-item__panel,
  .faq-item__answer,
  .faq-item__chevron {
    transition: none;
  }
}
</style>

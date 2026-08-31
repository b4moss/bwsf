<script setup lang="ts">
import {
  faqListInjectionKey,
  type FaqListContext,
} from "~/utils/faqListContext";

const { t } = useI18n();
const panelIds = ref<string[]>([]);
const openIds = ref<Set<string>>(new Set());

const allOpen = computed(
  () =>
    panelIds.value.length > 0 &&
    panelIds.value.every((id) => openIds.value.has(id)),
);

function registerPanel(id: string) {
  if (!panelIds.value.includes(id)) {
    panelIds.value = [...panelIds.value, id];
  }
}

function unregisterPanel(id: string) {
  panelIds.value = panelIds.value.filter((entry) => entry !== id);
  if (openIds.value.has(id)) {
    const next = new Set(openIds.value);
    next.delete(id);
    openIds.value = next;
  }
}

function isOpen(id: string) {
  return openIds.value.has(id);
}

function toggle(id: string) {
  const next = new Set(openIds.value);
  if (next.has(id)) {
    next.delete(id);
  } else {
    next.add(id);
  }
  openIds.value = next;
}

function toggleAll() {
  if (allOpen.value) {
    openIds.value = new Set();
  } else {
    openIds.value = new Set(panelIds.value);
  }
}

provide<FaqListContext>(faqListInjectionKey, {
  registerPanel,
  unregisterPanel,
  isOpen,
  toggle,
});
</script>

<template>
  <div class="faq-list">
    <div class="faq-list__controls">
      <button
        class="faq-list__switch"
        type="button"
        role="switch"
        :aria-checked="allOpen"
        :aria-label="allOpen ? t('faq.collapseAll') : t('faq.expandAll')"
        @click="toggleAll"
      >
        <span class="faq-list__switch-label">
          {{ allOpen ? t("faq.collapseAll") : t("faq.expandAll") }}
        </span>
        <span class="faq-list__switch-track" aria-hidden="true">
          <span class="faq-list__switch-thumb" />
        </span>
      </button>
    </div>
    <div class="faq-list__items">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.faq-list {
  margin: 1.5rem 0;
}

.faq-list__controls {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 0.85rem;
}

.faq-list__switch {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  margin: 0;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--color-muted);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
}

.faq-list__switch:hover {
  color: var(--color-ink);
}

.faq-list__switch-track {
  position: relative;
  width: 2.5rem;
  height: 1.35rem;
  border-radius: 999px;
  background: var(--color-border);
  transition: background-color 0.22s ease;
  flex-shrink: 0;
}

.faq-list__switch[aria-checked="true"] .faq-list__switch-track {
  background: var(--color-accent);
}

.faq-list__switch-thumb {
  position: absolute;
  top: 0.15rem;
  left: 0.15rem;
  width: 1.05rem;
  height: 1.05rem;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.12);
  transition: transform 0.22s ease;
}

.faq-list__switch[aria-checked="true"] .faq-list__switch-thumb {
  transform: translateX(1.15rem);
}

.faq-list__items {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
</style>

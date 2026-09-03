<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string;
    triggerText: string;
    compactText?: string;
  }>(),
  {
    compactText: "",
  },
);

const open = ref(false);
const root = ref<HTMLElement | null>(null);
const triggerId = useId();
const listId = useId();

function toggle() {
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function onDocumentPointerDown(event: PointerEvent) {
  const target = event.target as Node | null;
  if (!root.value || !target || root.value.contains(target)) {
    return;
  }
  close();
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    close();
  }
}

watch(open, (isOpen) => {
  if (!import.meta.client) {
    return;
  }
  if (isOpen) {
    document.addEventListener("pointerdown", onDocumentPointerDown);
    document.addEventListener("keydown", onKeydown);
  } else {
    document.removeEventListener("pointerdown", onDocumentPointerDown);
    document.removeEventListener("keydown", onKeydown);
  }
});

onBeforeUnmount(() => {
  if (!import.meta.client) {
    return;
  }
  document.removeEventListener("pointerdown", onDocumentPointerDown);
  document.removeEventListener("keydown", onKeydown);
});

defineExpose({ close, open });
</script>

<template>
  <div ref="root" class="dropdown" :data-open="open ? 'true' : 'false'">
    <button
      :id="triggerId"
      type="button"
      class="dropdown-trigger"
      :aria-label="label"
      :aria-expanded="open"
      aria-haspopup="listbox"
      :aria-controls="listId"
      @click="toggle"
    >
      <span class="trigger-text trigger-text--full">{{ triggerText }}</span>
      <span class="trigger-text trigger-text--compact">
        {{ compactText || triggerText }}
      </span>
      <span class="chevron" aria-hidden="true" />
    </button>
    <div
      v-show="open"
      :id="listId"
      class="dropdown-menu"
      role="listbox"
      :aria-labelledby="triggerId"
    >
      <slot :close="close" />
    </div>
  </div>
</template>

<style scoped>
.dropdown {
  position: relative;
}

.dropdown-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-height: 2.25rem;
  padding: 0.35rem 0.55rem;
  border: 1px solid var(--color-border);
  border-radius: 0.35rem;
  background: var(--color-surface);
  color: var(--color-ink);
  font: inherit;
  font-size: 0.875rem;
  line-height: 1.2;
  cursor: pointer;
}

.dropdown-trigger:hover {
  background: var(--color-accent-soft);
}

.dropdown[data-open="true"] .dropdown-trigger {
  border-color: color-mix(in srgb, var(--color-accent) 45%, var(--color-border));
}

.trigger-text--compact {
  display: none;
}

.chevron {
  width: 0.4rem;
  height: 0.4rem;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  transform: translateY(-1px) rotate(45deg);
  opacity: 0.7;
}

.dropdown[data-open="true"] .chevron {
  transform: translateY(1px) rotate(-135deg);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 0.35rem);
  right: 0;
  z-index: 40;
  min-width: max(100%, 11rem);
  padding: 0.3rem;
  border: 1px solid var(--color-border);
  border-radius: 0.45rem;
  background: var(--color-surface);
  box-shadow: 0 10px 28px color-mix(in srgb, var(--color-ink) 12%, transparent);
}

@media (max-width: 640px) {
  .dropdown-trigger {
    min-width: 2.25rem;
    justify-content: center;
    padding: 0.35rem 0.45rem;
  }

  .trigger-text--full {
    display: none;
  }

  .trigger-text--compact {
    display: inline;
    font-size: 0.8rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .dropdown-menu {
    min-width: 11rem;
    padding: 0.4rem;
  }
}
</style>

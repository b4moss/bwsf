<script setup lang="ts">
const { t } = useI18n();
const expanded = ref(false);

function toggle() {
  expanded.value = !expanded.value;
}
</script>

<template>
  <div class="panel collapse-box" :data-expanded="expanded">
    <div class="collapse-box__body">
      <slot />
    </div>
    <button
      class="btn collapse-box__toggle"
      type="button"
      :aria-expanded="expanded"
      @click="toggle"
    >
      {{
        expanded
          ? t("collapseBox.collapse")
          : t("collapseBox.expand")
      }}
    </button>
  </div>
</template>

<style scoped>
.collapse-box {
  margin: 1.25rem 0;
  padding-bottom: 0.85rem;
}

.collapse-box__body {
  position: relative;
  max-height: 6rem;
  overflow: hidden;
}

.collapse-box[data-expanded="true"] .collapse-box__body {
  max-height: none;
  overflow: visible;
}

.collapse-box[data-expanded="false"] .collapse-box__body::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2.75rem;
  pointer-events: none;
  background: linear-gradient(
    to bottom,
    transparent,
    var(--color-surface)
  );
}

.collapse-box__toggle {
  margin-top: 0.75rem;
  background: transparent;
  color: var(--color-accent);
  border: 1px solid var(--color-border);
  font-weight: 500;
  padding: 0.4rem 0.85rem;
}

.collapse-box__toggle:hover {
  background: var(--color-accent-soft);
  color: var(--color-accent-hover);
}
</style>

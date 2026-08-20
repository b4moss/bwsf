<script setup lang="ts">
const { t } = useI18n();
const analyticsEnabled = ref(true);
const saved = ref(false);

onMounted(() => {
  const consent = localStorage.getItem("cookie-consent");
  analyticsEnabled.value = consent !== "declined";
});

function saveSettings() {
  localStorage.setItem(
    "cookie-consent",
    analyticsEnabled.value ? "accepted" : "declined",
  );
  saved.value = true;
  setTimeout(() => {
    saved.value = false;
  }, 2000);
}
</script>

<template>
  <div class="cookie-settings">
    <h2>{{ t("cookie.settingsTitle") }}</h2>

    <div class="setting-item">
      <div class="setting-header">
        <label class="setting-label">{{ t("cookie.analyticsLabel") }}</label>
        <div class="toggle-wrapper">
          <button
            type="button"
            :class="['toggle-button', { active: analyticsEnabled }]"
            :aria-pressed="analyticsEnabled"
            @click="analyticsEnabled = !analyticsEnabled"
          >
            <span class="toggle-slider" />
          </button>
          <span class="toggle-status">
            {{ analyticsEnabled ? t("cookie.enabled") : t("cookie.disabled") }}
          </span>
        </div>
      </div>
      <p class="setting-description">{{ t("cookie.analyticsDescription") }}</p>
    </div>

    <div class="setting-actions">
      <button type="button" class="save-button" @click="saveSettings">
        {{ saved ? t("cookie.saved") : t("cookie.save") }}
      </button>
    </div>

    <p class="setting-note">{{ t("cookie.note") }}</p>
  </div>
</template>

<style scoped>
.cookie-settings {
  margin: 2rem 0;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
}

.cookie-settings h2 {
  margin: 0 0 1.5rem 0;
  padding: 0;
  border: none;
  font-size: 1.25rem;
}

.setting-item {
  padding: 1rem 0;
  border-bottom: 1px solid var(--color-border);
}

.setting-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.setting-label {
  font-weight: 600;
  color: var(--color-ink);
}

.toggle-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.toggle-button {
  position: relative;
  width: 48px;
  height: 26px;
  border-radius: 13px;
  border: none;
  background: var(--color-border);
  cursor: pointer;
  transition: background 0.2s ease;
}

.toggle-button.active {
  background: var(--color-accent);
}

.toggle-slider {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: white;
  transition: transform 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.toggle-button.active .toggle-slider {
  transform: translateX(22px);
}

.toggle-status {
  font-size: 0.875rem;
  color: var(--color-muted);
  min-width: 50px;
}

.setting-description {
  margin: 0.5rem 0 0 0;
  font-size: 0.875rem;
  color: var(--color-muted);
}

.setting-actions {
  margin-top: 1.5rem;
}

.save-button {
  padding: 0.625rem 1.25rem;
  border-radius: 6px;
  border: none;
  background: var(--color-accent);
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s ease;
}

.save-button:hover {
  background: var(--color-accent-hover);
}

.setting-note {
  margin: 1rem 0 0 0;
  font-size: 0.8rem;
  color: var(--color-muted);
}
</style>

<script setup lang="ts">
const { t } = useI18n();
const localePath = useLocalePath();
const showBanner = ref(false);
const { loadGoogleAnalytics } = useGoogleAnalytics();

const policyPath = computed(() => localePath("/cookie-policy"));

onMounted(() => {
  const consent = localStorage.getItem("cookie-consent");
  if (!consent) {
    showBanner.value = true;
  }
});

function acceptCookies() {
  localStorage.setItem("cookie-consent", "accepted");
  showBanner.value = false;
  loadGoogleAnalytics();
  const w = window as Window & { gtag?: (...args: unknown[]) => void };
  if (w.gtag) {
    w.gtag("consent", "update", { analytics_storage: "granted" });
  }
}

function declineCookies() {
  localStorage.setItem("cookie-consent", "declined");
  showBanner.value = false;
  const w = window as Window & { gtag?: (...args: unknown[]) => void };
  if (w.gtag) {
    w.gtag("consent", "update", { analytics_storage: "denied" });
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="slide-up">
      <div v-if="showBanner" class="cookie-consent">
        <div class="cookie-consent-content">
          <p class="cookie-consent-message">{{ t("cookie.message") }}</p>
          <div class="cookie-consent-actions">
            <NuxtLink :to="policyPath" class="cookie-consent-link">
              {{ t("cookie.learnMore") }}
            </NuxtLink>
            <button
              type="button"
              class="cookie-consent-button decline"
              @click="declineCookies"
            >
              {{ t("cookie.decline") }}
            </button>
            <button
              type="button"
              class="cookie-consent-button accept"
              @click="acceptCookies"
            >
              {{ t("cookie.accept") }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.cookie-consent {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  box-shadow: 0 -4px 12px color-mix(in srgb, var(--color-ink) 12%, transparent);
}

.cookie-consent-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
}

.cookie-consent-message {
  margin: 0;
  font-size: 14px;
  color: var(--color-ink);
  flex: 1;
  min-width: 200px;
}

.cookie-consent-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.cookie-consent-link {
  font-size: 14px;
  color: var(--color-accent);
  text-decoration: none;
  white-space: nowrap;
}

.cookie-consent-link:hover {
  text-decoration: underline;
}

.cookie-consent-button {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.cookie-consent-button.decline {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-muted);
}

.cookie-consent-button.decline:hover {
  border-color: var(--color-muted);
  color: var(--color-ink);
}

.cookie-consent-button.accept {
  background: var(--color-accent);
  border: 1px solid var(--color-accent);
  color: #fff;
}

.cookie-consent-button.accept:hover {
  background: var(--color-accent-hover);
  border-color: var(--color-accent-hover);
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition:
    transform 0.3s ease,
    opacity 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

@media (max-width: 640px) {
  .cookie-consent-content {
    flex-direction: column;
    align-items: stretch;
    text-align: center;
  }

  .cookie-consent-actions {
    justify-content: center;
  }
}
</style>

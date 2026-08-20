<template>
  <div class="site-shell">
    <SiteHeader />
    <div class="site-body">
      <DocsSidebar />
      <div class="site-content">
        <main class="site-main">
          <slot />
        </main>
        <SiteFooter />
      </div>
    </div>
  </div>
</template>

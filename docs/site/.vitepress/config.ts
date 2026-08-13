import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'bwsf',
  description: 'Manage .env* and Terraform tfvars with Bitwarden CLI',
  
  // Custom domain root (https://bwsf.oss.b4m.jp/)
  base: '/',
  
  // Clean URLs (no .html extension)
  cleanUrls: true,
  
  // Head meta tags
  head: [
    ['meta', { name: 'theme-color', content: '#175ddc' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:site_name', content: 'bwsf' }],
    // Google Analytics is loaded dynamically based on cookie consent
    // See .vitepress/theme/useGoogleAnalytics.ts
  ],
  
  // i18n configuration
  locales: {
    en: {
      label: 'English',
      lang: 'en',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/getting-started' },
          { text: 'Commands', link: '/en/guide/commands' },
          {
            text: 'Links',
            items: [
              { text: 'GitHub', link: 'https://github.com/b4moss/bwsf' },
              { text: 'Changelog', link: 'https://github.com/b4moss/bwsf/releases' },
              { text: 'Our Company', link: 'https://b4m.co.jp/' },
            ]
          }
        ],
        sidebar: [
          {
            text: 'Introduction',
            items: [
              { text: 'What is bwsf?', link: '/en/guide/getting-started' },
              { text: 'Key Features', link: '/en/guide/features' },
            ]
          },
          {
            text: 'Usage',
            items: [
              { text: 'Installation', link: '/en/guide/installation' },
              { text: 'Commands', link: '/en/guide/commands' },
              { text: 'Philosophy', link: '/en/guide/philosophy' },
              {
                text: 'Other',
                collapsed: true,
                items: [
                  { text: 'Upgrade', link: '/en/guide/upgrade' },
                  { text: 'Uninstall', link: '/en/guide/uninstall' },
                  { text: 'Features in Development', link: '/en/guide/dev-loadmap' },
                ]
              },
              { text: 'FAQ', link: '/en/guide/faq' },
            ]
          },
          {
            text: 'Legal',
            items: [
              { text: 'License', link: '/en/guide/license' },
              { text: 'License FAQ', link: '/en/guide/license-faq' },
              { text: 'Cookie Policy', link: '/en/cookie-policy' },
            ]
          }
        ],
        footer: {
          message: 'Released under the MIT License.',
          copyright: 'Copyright © b4m-oss, Bicycle for Mind LLC.'
        },
        editLink: {
          pattern: 'https://github.com/b4moss/bwsf/edit/main/docs/site/:path',
          text: 'Edit this page on GitHub'
        }
      }
    },
    ja: {
      label: '日本語',
      lang: 'ja',
      link: '/ja/',
      themeConfig: {
        nav: [
          { text: 'ガイド', link: '/ja/guide/getting-started' },
          { text: 'コマンド', link: '/ja/guide/commands' },
          {
            text: 'リンク',
            items: [
              { text: 'GitHub', link: 'https://github.com/b4moss/bwsf' },
              { text: '変更履歴', link: 'https://github.com/b4moss/bwsf/releases' },
              { text: '合同会社 知的・自転車', link: 'https://b4m.co.jp/' },
            ]
          }
        ],
        sidebar: [
          {
            text: 'はじめに',
            items: [
              { text: 'bwsfとは？', link: '/ja/guide/getting-started' },
              { text: '主な機能', link: '/ja/guide/features' },
            ]
          },
          {
            text: '使い方',
            items: [
              { text: 'インストール', link: '/ja/guide/installation' },
              { text: 'コマンド', link: '/ja/guide/commands' },
              { text: 'フィロソフィー', link: '/ja/guide/philosophy' },
              {
                text: 'その他',
                collapsed: true,
                items: [
                  { text: 'アップグレード', link: '/ja/guide/upgrade' },  
                  { text: 'アンインストール', link: '/ja/guide/uninstall' },
                  { text: '開発予定中の機能', link: '/ja/guide/dev-loadmap' },
                ]
              },
              { text: 'よくある質問', link: '/ja/guide/faq' },
            ]
          },
          {
            text: '法的事項',
            items: [
              { text: 'ライセンス', link: '/ja/guide/license' },
              { text: 'ライセンスに関するFAQ', link: '/ja/guide/license-faq' },
              { text: 'Cookieポリシー', link: '/ja/cookie-policy' },
            ]
          }
        ],
        footer: {
          message: 'MITライセンスの下で公開されています。',
          copyright: 'Copyright © b4m-oss, 合同会社 知的・自転車'
        },
        editLink: {
          pattern: 'https://github.com/b4moss/bwsf/edit/main/docs/site/:path',
          text: 'GitHubでこのページを編集'
        },
        docFooter: {
          prev: '前のページ',
          next: '次のページ'
        },
        outline: {
          label: '目次'
        },
        lastUpdated: {
          text: '最終更新'
        },
        returnToTopLabel: 'トップに戻る',
        sidebarMenuLabel: 'メニュー',
        darkModeSwitchLabel: 'ダークモード',
      }
    }
  },
  
  themeConfig: {
    // Logo and site title
    logo: '/logo.svg',
    siteTitle: 'bwsf',
    
    // Social links
    socialLinks: [
      { icon: 'github', link: 'https://github.com/b4moss/bwsf' }
    ],
    
    // Search
    search: {
      provider: 'local'
    }
  }
})

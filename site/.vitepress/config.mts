import { defineConfig } from 'vitepress'

const repo = 'https://github.com/prea-framework/prea'

export default defineConfig({
  title: 'Preá',
  description: 'A modular web framework for Go.',
  base: '/prea/',
  lang: 'en-US',
  cleanUrls: true,
  lastUpdated: true,
  srcExclude: ['STYLE.md', 'examples/**'],

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/prea/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#2B1F17' }],
    ['meta', { property: 'og:title', content: 'Preá — a modular web framework for Go' }],
    ['meta', { property: 'og:image', content: '/prea/logo-full.svg' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Lexend:wght@600;700&display=swap' }],
  ],

  markdown: {
    theme: { light: 'github-light', dark: 'github-dark' },
  },

  themeConfig: {
    logo: { light: '/logo-light.svg', dark: '/logo.svg', alt: '' },
    siteTitle: 'preá',

    nav: [
      { text: 'Docs', link: '/guide/what-is-prea', activeMatch: '/guide/' },
      { text: 'Tutorial', link: '/tutorial/', activeMatch: '/tutorial/' },
      { text: 'Reference', link: '/reference/cli', activeMatch: '/reference/' },
      { text: 'Persistence', link: '/persistence/', activeMatch: '/persistence/' },
      { text: 'About', link: '/about/the-name', activeMatch: '/about/' },
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Getting started',
          items: [
            { text: 'What is Preá?', link: '/guide/what-is-prea' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Quick start', link: '/guide/quick-start' },
            { text: 'Project structure', link: '/guide/project-structure' },
          ],
        },
        {
          text: 'Fundamentals',
          items: [
            { text: 'Preás', link: '/guide/preas' },
            { text: 'Components & injection', link: '/guide/dependency-injection' },
            { text: 'Handlers', link: '/guide/handlers' },
            { text: 'Services', link: '/guide/services' },
            { text: 'Errors & responses', link: '/guide/errors-and-responses' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Middlewares', link: '/guide/middlewares' },
            { text: 'App lifecycle', link: '/guide/lifecycle' },
          ],
        },
        {
          text: 'Database',
          items: [
            { text: 'Models & schema', link: '/guide/models' },
            { text: 'Migrations', link: '/guide/migrations' },
            { text: 'Repositories', link: '/guide/repositories' },
            { text: 'Transactions', link: '/guide/transactions' },
          ],
        },
        {
          text: 'Guides',
          items: [
            { text: 'HTTP security', link: '/guide/http-security' },
            { text: 'Tokens & passwords', link: '/guide/tokens-and-passwords' },
            { text: 'Logging', link: '/guide/logging' },
            { text: 'Testing', link: '/guide/testing' },
            { text: 'Deploying', link: '/guide/deploying' },
          ],
        },
      ],
      '/tutorial/': [
        {
          text: 'Build a todo API',
          items: [
            { text: 'Introduction', link: '/tutorial/' },
            { text: '1. Create the project', link: '/tutorial/create-the-project' },
            { text: '2. Your first preá', link: '/tutorial/first-prea' },
            { text: '3. The model', link: '/tutorial/the-model' },
            { text: '4. Business rules', link: '/tutorial/business-rules' },
            { text: '5. A second preá', link: '/tutorial/second-prea' },
            { text: '6. A middleware', link: '/tutorial/middleware' },
            { text: '7. Configuration', link: '/tutorial/configuration' },
            { text: '8. Testing', link: '/tutorial/testing' },
            { text: '9. Ship it', link: '/tutorial/ship-it' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'CLI', link: '/reference/cli' },
            { text: 'Configuration keys', link: '/reference/configuration' },
            { text: 'Struct tags', link: '/reference/struct-tags' },
            { text: 'Errors & HTTP status', link: '/reference/errors' },
            { text: 'Go API', link: '/reference/api' },
          ],
        },
      ],
      '/about/': [
        {
          text: 'About',
          items: [
            { text: 'The name', link: '/about/the-name' },
            { text: "What's next", link: '/about/whats-next' },
            { text: 'Changelog', link: '/about/changelog' },
            { text: 'Contributing', link: '/about/contributing' },
          ],
        },
      ],
    },

    outline: { level: [2, 3] },
    search: { provider: 'local' },
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: {
      pattern: `${repo}/edit/main/site/:path`,
      text: 'Edit this page on GitHub',
    },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Made in the Brazilian Northeast.',
    },
  },
})

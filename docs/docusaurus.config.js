// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import {themes as prismThemes} from 'prism-react-renderer';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

// GitHub Pages: `/mb-cli/`. Para pré-visualizar `dist` com `serve` na raiz (ex.: `make docs-preview`), use
// `DOCUSAURUS_BASE_URL=/ npm run build` — ver Makefile.
const baseUrl = process.env.DOCUSAURUS_BASE_URL ?? '/mb-cli/';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'MB CLI',
  tagline: 'CLI em Go para orquestrar plugins com injeção segura de variáveis de ambiente.',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://carlosdorneles-mb.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served (GitHub Pages: /<projectName>/)
  baseUrl,

  // GitHub pages deployment config (ajuste para o seu org/repo).
  organizationName: 'mercadobitcoin',
  projectName: 'mb-cli',

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  markdown: {
    mermaid: true,
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: './content',
          sidebarPath: './sidebars.js',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themes: ['@docusaurus/theme-mermaid', '@easyops-cn/docusaurus-search-local'],

  plugins: [
    '@r74tech/docusaurus-plugin-panzoom',
    function webpackIgnoreWarnings() {
      return {
        name: 'webpack-ignore-critical-dep',
        configureWebpack() {
          return {
            ignoreWarnings: [
              /Critical dependency: require function is used in a way/,
            ],
          };
        },
      };
    },
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      image: 'img/docusaurus-social-card.jpg',
      colorMode: {
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'MB CLI',
        logo: {
          alt: 'MB CLI Logo',
          src: 'img/logo.png',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docsSidebar',
            position: 'left',
            label: 'Documentação',
          },
          {
            href: 'https://github.com/carlosdorneles-mb/mb-cli',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Documentação',
            items: [
              { label: 'Introdução', to: '/docs/intro' },
              { label: 'Começar', to: '/docs/getting-started/' },
              { label: 'Criar um plugin', to: '/docs/plugin-authoring/create-a-plugin' },
            ],
          },
          {
            title: 'Mais',
            items: [
              { label: 'GitHub', href: 'https://github.com/carlosdorneles-mb/mb-cli' },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} MB CLI.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
      // Pan/zoom nos diagramas Mermaid (e SVG). Inclui controles de zoom in/out/reset.
      zoom: {
        selectors: ['div.mermaid[data-processed="true"]', 'div.docusaurus-mermaid-container'],
        wrap: true,
        timeout: 1000,
      },
    }),
};

export default config;

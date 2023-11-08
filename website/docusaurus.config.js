// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'GO Feature Flag',
  tagline: 'Simple Feature Flagging for All',
  url: 'https://gofeatureflag.org',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon/favicon.png',
  organizationName: 'thomaspoignant',
  projectName: 'go-feature-flag',
  trailingSlash: false,

  customFields: {
    description: 'GO Feature Flag is a simple, complete and lightweight feature flag solution 100% Open Source.',
    logo: 'img/logo/logo.png',
    github: 'https://github.com/thomaspoignant/go-feature-flag',
    sponsor: 'https://github.com/sponsors/thomaspoignant',
    openfeature: 'https://openfeature.dev',
    mailchimpURL:
      '//gofeatureflag.us14.list-manage.com/subscribe/post?u=86acc1a78e371bf66a9683672&amp;id=f42abfec51&amp',
    swaggerURL:
      'https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/cmd/relayproxy/docs/swagger.yaml',
    playgroundEvaluationApi: 'https://fjaf6mppiu.eu-west-1.awsapprunner.com/v1/feature/evaluate',
  },
  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        googleAnalytics: {
          trackingID: 'G-LEJBB94YBE',
        },
        gtag: {
          trackingID: 'G-LEJBB94YBE',
        },
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/thomaspoignant/go-feature-flag/tree/main/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/thomaspoignant/go-feature-flag/tree/main/website/',
        },
        theme: {
          customCss: [
            require.resolve('./src/css/custom.css'),
            require.resolve('./src/css/pushy-buttons.css'), //https://github.com/iRaul/pushy-buttons
            require.resolve('./src/css/simplegrid.css'), //https://thisisdallas.github.io/Simple-Grid/
          ],
        },
        sitemap: {
          changefreq: 'weekly',
          priority: 0.5,
          ignorePatterns: ['/tags/**'],
          filename: 'sitemap.xml',
        }
      }),
    ],
  ],
  stylesheets: [
    'https://fonts.googleapis.com/css?family=Poppins:400,500,700',
    'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.2.0/css/all.min.css',
    'https://cdn.jsdelivr.net/gh/devicons/devicon@v2.15.1/devicon.min.css', // https://devicon.dev/
  ],
  themes: [
    [
      '@easyops-cn/docusaurus-search-local',
      /** @type {import("@easyops-cn/docusaurus-search-local").PluginOptions} */
      ({
        hashed: true,
        language: ['en', 'zh'],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      }),
    ],
  ],
  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      announcementBar: {
        id: 'support_usz', // Increment on change
        content: `⭐ If you like GO Feature Flag, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/thomaspoignant/go-feature-flag">GitHub</a> and follow us on <a target="_blank" rel="noopener noreferrer" href="https://x.com/gofeatureflag">X</a>`,
      },
      image: 'img/logo/x-card.png',
      navbar: {
        title: 'GO Feature Flag',
        logo: {
          alt: 'GO Feature Flag Logo',
          src: 'img/logo/logo_128.png',
        },
        items: [
          {
            position: 'left',
            label: 'Product',
            items: [
              {
                to: '/product/what_is_feature_management',
                html: '<i class="fa-solid fa-list-check menu-icon"></i> What is Feature Management?',
              },
              {
                to: '/product/why_go_feature_flag',
                html: '<i class="fa-solid fa-laptop-code menu-icon"></i> Why GO Feature Flag?',
              },
              {
                to: '/product/open_feature_support',
                html: '<i class="fa-solid fa-toggle-on menu-icon"></i> Open Feature Support',
              }
            ]
          },
          {
            position: 'left',
            label: 'Developers',
            items: [
              {
                to: '/docs/category/getting-started',
                html: '<i class="fa-solid fa-rocket menu-icon"></i> Getting Started',
              },
              {
                to: '/docs/openfeature_sdk/sdk',
                html: '<i class="fa-solid fa-code menu-icon"></i> SDKs',
              },
              {
                to: '/editor',
                html: '<i class="fa-solid fa-pencil menu-icon"></i> Flag Editor',
              },
              {
                html: '<i class="fa-solid fa-book menu-icon"></i> Documentation',
                type: 'doc',
                docId: 'index'
              },
              {
                html: '<i class="fa-solid fa-eye menu-icon"></i> Examples',
                to: 'https://github.com/thomaspoignant/go-feature-flag/tree/main/examples',
              },
              {
                html: '<i class="fa-solid fa-star menu-icon"></i> Feature Flag Best Practice',
                to: '/blog/feature-flag-best-practice'
              },
              {
                to: 'https://gophers.slack.com/messages/go-feature-flag',
                html: '<i class="fa-brands fa-slack menu-icon"></i> Community <i class="fa fa-external-link" aria-hidden="true"></i>',
              },
              {
                to: 'https://github.com/thomaspoignant/go-feature-flag/releases',
                html: '<i class="fa-brands fa-github menu-icon"></i> Changelog <i class="fa fa-external-link" aria-hidden="true"></i>',
              }
            ]
          },
          {type: 'doc', docId: 'index', position: 'left', html: 'Docs'},
          {to: '/blog', label: 'Blog', position: 'left'},
          {to: '/editor', html: 'Editor', position: 'left'},
          {
            to: 'https://github.com/sponsors/thomaspoignant',
            label: 'Sponsor us ❤️',
            position: 'right',
          },
          {
            type: 'search',
            position: 'right',
          },
          {
            type: 'docsVersionDropdown',
            position: 'right',
            dropdownItemsAfter: [{to: '/versions', label: 'All versions'}],
          },
          {
            href: 'https://github.com/thomaspoignant/go-feature-flag',
            position: 'right',
            className: 'header-github-link navbar__right',
            'aria-label': 'GitHub repository',
            dropdownItemsAfter: [{to: '/versions', label: 'All versions'}],
          },
          {
            href: 'https://x.com/gofeatureflag',
            position: 'right',
            className: 'header-twitter-link navbar__right',
            'aria-label': 'Twitter',
          },
          {
            href: 'https://gophers.slack.com/messages/go-feature-flag',
            position: 'right',
            className: 'header-slack-link navbar__right',
            'aria-label': 'Slack',
          }
        ],
      },
      footer: {
        logo: {
          alt: 'GO Feature Flag logo',
          src: 'img/logo/logo.png',
          href: 'https://gofeatureflag.org',
          width: 100,
        },
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Getting Started',
                to: '/docs/category/getting-started',
              },
              {
                label: 'GO Module',
                to: '/docs/category/use-as-a-go-module'
              },
              {
                label: 'SDKs',
                to: '/docs/openfeature_sdk/sdk',
              },
              {
                label: 'Relay Proxy',
                to: '/docs/relay_proxy'
              }
            ],
          },

          {
            title: 'Community',
            items: [
              {
                html: `
                <a href="https://gophers.slack.com/messages/go-feature-flag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-slack"></i> Slack&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://x.com/gofeatureflag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-x-twitter"></i> X&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://youtube.com/@gofeatureflag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-youtube"></i> Youtube &nbsp;<i class="fa-solid fa-arrow-up-right-from-square"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://github.com/thomaspoignant/go-feature-flag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-github"></i> Github&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square"></i>
                </a>`,
              },
              {
                html: `
                <a href="mailto:contact@gofeatureflag.org" target="_blank" rel="noreferrer noopener">
                  <i class="fa-regular fa-envelope"></i> Email&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square"></i>
                </a>`,
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Blog',
                to: '/blog',
              }
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} GO Feature Flag.`,
      },
      prism: {
        theme: require("prism-react-renderer").themes.vsLight,
        darkTheme: require("prism-react-renderer").themes.vsDark,
        additionalLanguages: ['java', 'scala', 'toml', 'php', 'go', 'csharp', 'yaml', 'python'],
      },
    }),
};

module.exports = config;

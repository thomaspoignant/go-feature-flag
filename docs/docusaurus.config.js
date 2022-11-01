// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');
const TwitterSvg =
  '<svg style="fill: #1DA1F2; vertical-align: middle; margin-left: 3px;" width="16" height="16" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path d="M459.37 151.716c.325 4.548.325 9.097.325 13.645 0 138.72-105.583 298.558-298.558 298.558-59.452 0-114.68-17.219-161.137-47.106 8.447.974 16.568 1.299 25.34 1.299 49.055 0 94.213-16.568 130.274-44.832-46.132-.975-84.792-31.188-98.112-72.772 6.498.974 12.995 1.624 19.818 1.624 9.421 0 18.843-1.3 27.614-3.573-48.081-9.747-84.143-51.98-84.143-102.985v-1.299c13.969 7.797 30.214 12.67 47.431 13.319-28.264-18.843-46.781-51.005-46.781-87.391 0-19.492 5.197-37.36 14.294-52.954 51.655 63.675 129.3 105.258 216.365 109.807-1.624-7.797-2.599-15.918-2.599-24.04 0-57.828 46.782-104.934 104.934-104.934 30.213 0 57.502 12.67 76.67 33.137 23.715-4.548 46.456-13.32 66.599-25.34-7.798 24.366-24.366 44.833-46.132 57.827 21.117-2.273 41.584-8.122 60.426-16.243-14.292 20.791-32.161 39.308-52.628 54.253z"></path></svg>';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'GO Feature Flag',
  tagline: 'Ship Faster, Reduce Risk, and Build Scale.',
  url: 'https://docs.gofeatureflag.org',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon/favicon.png',
  organizationName: 'thomaspoignant',
  projectName: 'go-feature-flag',
  trailingSlash: false,

  customFields:{
    logo: "img/logo/logo.png",
    github: "https://github.com/thomaspoignant/go-feature-flag",
    sponsor: "https://github.com/sponsors/thomaspoignant",
    openfeature: "https://openfeature.dev",
    mailchimpURL: "//gofeatureflag.us14.list-manage.com/subscribe/post?u=86acc1a78e371bf66a9683672&amp;id=f42abfec51&amp"
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
          trackingID: 'G-YMBZZ2GYSK',
          anonymizeIP: true,
        },
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/thomaspoignant/go-feature-flag/tree/main/packages/create-docusaurus/templates/shared/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/thomaspoignant/go-feature-flag/tree/main/packages/create-docusaurus/templates/shared/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],
  stylesheets: [
    'css/pushy-buttons.css', //https://github.com/iRaul/pushy-buttons
    'css/simplegrid.css',
    'https://fonts.googleapis.com/css?family=Poppins:400,500,700',
    'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.2.0/css/all.min.css'
  ],
  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      announcementBar: {
        id: 'support_usz', // Increment on change
        content: `⭐️ If you like GO Feature Flag, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/thomaspoignant/go-feature-flag">GitHub</a> and follow us on <a target="_blank" rel="noopener noreferrer" href="https://twitter.com/gofeatureflag">Twitter ${TwitterSvg}</a>`,
      },
      navbar: {
        title: 'GO Feature Flag',
        logo: {
          alt: 'GO Feature Flag Logo',
          src: 'img/logo/logo_128.png',
        },
        items: [
          {
            type: 'doc',
            docId: 'index',
            position: 'left',
            label: 'Docs',
          },
          {to: '/blog', label: 'Blog', position: 'left'},
          {to: 'https://editor.gofeatureflag.org', html: 'Flag Editor <i class="fas fa-external-link-alt"></i>', position: 'left'},
          {to: 'https://github.com/sponsors/thomaspoignant', label: 'Sponsor us ❤️', position: 'right'},
          {type: 'docsVersionDropdown', position: 'right'},
          {
            href: 'https://github.com/thomaspoignant/go-feature-flag',
            position: 'right',
            className: "header-github-link navbar__right",
            "aria-label": "GitHub repository",
          },
          {
            href: 'https://twitter.com/gofeatureflag',
            position: 'right',
            className: "header-twitter-link navbar__right",
            "aria-label": "Twitter",
          },
          {
            href: 'https://gophers.slack.com/messages/go-feature-flag',
            position: 'right',
            className: "header-slack-link navbar__right",
            "aria-label": "Slack",
          },
          {
            href: 'mailto:contact@gofeatureflag.org',
            position: 'right',
            className: "header-email-link navbar__right",
            "aria-label": "Email",
          },
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
            title: 'Documentation',
            items: [
              {
                label: 'Docs',
                to: '/docs/',
              },
            ],
          },

          {
            title: 'Community',
            items: [
              {
                label: 'Slack',
                href: 'https://gophers.slack.com/messages/go-feature-flag',
              },
              {
                label: 'Twitter',
                href: 'https://twitter.com/gofeatureflag',
              },
              {
                label: 'Github',
                href: 'https://github.com/thomaspoignant/go-feature-flag',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Blog',
                to: '/blog',
              },
              {
                label: 'GitHub',
                href: 'https://github.com/thomaspoignant/go-feature-flag',
              },
            ],
          },

        ],
        copyright: `Copyright © ${new Date().getFullYear()} GO Feature Flag.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
};

module.exports = config;

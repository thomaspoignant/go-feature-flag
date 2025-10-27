// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const {sdk} = require('./data/sdk');
const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;
const {generateSdksDropdownHTML} = require('./src/components/navbar/sdks');

/** @type {import("@docusaurus/types").Config} */
const config = {
  title: 'GO Feature Flag',
  tagline: 'Simple Feature Flagging for All',
  url: 'https://gofeatureflag.org',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  favicon: 'img/favicon/favicon.ico',
  organizationName: 'thomaspoignant',
  projectName: 'go-feature-flag',
  trailingSlash: false,
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
    mermaid: true,
  },
  plugins: [
    [
      '@docusaurus/plugin-client-redirects',
      {
        redirects: [
          {
            from: '/docs/configure_flag/flag_format',
            to: '/docs/configure_flag/create-flags',
          },
          {
            from: '/docs/configure_flag/rule_format',
            to: '/docs/configure_flag/target-with-flags',
          },
          {
            from: '/docs/configure_flag/rollout/experimentation',
            to: '/docs/configure_flag/rollout-strategies/experimentation',
          },
          {
            from: '/docs/go_module/store_file/kubernetes_configmaps',
            to: '/docs/integrations/store-flags-configuration/kubernetes-configmap',
          },
          {
            from: '/docs/getting_started/using-go-module',
            to: '/docs/go_module/getting-started',
          },
          {from: '/docs/openfeature_sdk/sdk', to: '/docs/sdk'},
          {from: '/docs/category/getting-started', to: '/docs/getting-started'},
          {
            from: '/docs/relay_proxy/configure_relay_proxy',
            to: '/docs/relay-proxy/configure-relay-proxy',
          },
          {
            from: '/docs/configure_flag/store_your_flags',
            to: '/docs/integrations/store-flags-configuration',
          },
          {
            from: '/docs/getting_started/using-openfeature',
            to: '/docs/getting-started',
          },
          {
            from: '/docs/configure_flag/rollout/progressive',
            to: '/docs/configure_flag/rollout-strategies/progressive',
          },
          {
            from: '/docs/category/configure-your-feature-flags',
            to: '/docs/configure_flag/create-flags',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_ruby',
            to: '/docs/sdk/server_providers/openfeature_ruby',
          },
          {
            from: '/docs/go_module/store_file/mongodb',
            to: '/docs/integrations/store-flags-configuration/mongodb',
          },
          {
            from: '/docs/relay_proxy/deploy_relay_proxy',
            to: '/docs/relay-proxy/deployment',
          },
          {
            from: '/docs/go_module/store_file/github',
            to: '/docs/integrations/store-flags-configuration/github',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_go',
            to: '/docs/sdk/server_providers/openfeature_go',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_java',
            to: '/docs/sdk/server_providers/openfeature_java',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_javascript',
            to: '/docs/sdk/server_providers/openfeature_javascript',
          },
          {
            from: '/docs/go_module/store_file/custom',
            to: '/docs/integrations/store-flags-configuration#custom-retriever',
          },
          {
            from: '/docs/go_module/store_file/http',
            to: '/docs/integrations/store-flags-configuration/http',
          },
          {
            from: '/docs/relay_proxy/getting_started',
            to: '/docs/relay-proxy/getting_started',
          },
          {
            from: '/docs/configure_flag/rollout/scheduled',
            to: '/docs/configure_flag/rollout-strategies/scheduled',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_python',
            to: '/docs/sdk/server_providers/openfeature_python',
          },
          {
            from: '/docs/relay_proxy/advanced_usage',
            to: '/docs/relay-proxy/advanced_usage',
          },
          {
            from: '/docs/relay_proxy/relay_proxy_endpoints',
            to: '/docs/relay-proxy/relay_proxy_endpoints',
          },
          {
            from: '/docs/go_module/store_file/file',
            to: '/docs/integrations/store-flags-configuration/file',
          },
          {
            from: '/docs/openfeature_sdk/client_providers/openfeature_react',
            to: '/docs/sdk/client_providers/openfeature_react',
          },
          {
            from: '/docs/relay_proxy/monitor_relay_proxy',
            to: '/docs/relay-proxy/observability',
          },
          {
            from: '/docs/configure_flag/export_flags_usage',
            to: '/docs/integrations/export-evaluation-data',
          },
          {
            from: '/docs/experimental/ofrep',
            to: '/API_relayproxy#tag/OpenFeature-Remote-Evaluation-Protocol-(OFREP)',
          },
          {
            from: '/docs/go_module/data_collection/s3',
            to: '/docs/integrations/export-evaluation-data/aws-s3',
          },
          {
            from: '/docs/go_module/notifier/slack',
            to: '/docs/integrations/notify-flags-changes/slack',
          },
          {
            from: '/docs/go_module/notifier/webhook',
            to: '/docs/integrations/notify-flags-changes/webhook',
          },
          {
            from: '/docs/go_module/store_file/google_cloud_storage',
            to: '/docs/integrations/store-flags-configuration/google-cloud-storage',
          },
          {
            from: '/docs/go_module/store_file/redis',
            to: '/docs/integrations/store-flags-configuration/redis',
          },
          {
            from: '/docs/go_module/store_file/s3',
            to: '/docs/integrations/store-flags-configuration/aws-s3',
          },
          {
            from: '/docs/next/configure_flag/rollout/scheduled',
            to: '/docs/configure_flag/rollout-strategies/scheduled',
          },
          {
            from: '/docs/openfeature_sdk/client_providers/openfeature_javascript',
            to: '/docs/sdk/client_providers/openfeature_javascript',
          },
          {
            from: '/docs/openfeature_sdk/client_providers/openfeature_swift',
            to: '/docs/sdk/client_providers/openfeature_swift',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_dotnet',
            to: '/docs/sdk/server_providers/openfeature_dotnet',
          },
          {
            from: '/docs/openfeature_sdk/server_providers/openfeature_php',
            to: '/docs/sdk/server_providers/openfeature_php',
          },
          {
            from: '/docs/relay_proxy/install_relay_proxy',
            to: '/docs/relay-proxy/install_relay_proxy',
          },
        ],
      },
    ],
    require('./plugins/tailwind-plugin.cjs'),
  ],

  customFields: {
    description:
      'GO Feature Flag is a simple, complete and lightweight feature flag solution 100% Open Source. Get the full feature flag experience using OpenFeature and GO Feature Flag.',
    logo: 'img/logo/logo.png',
    github: 'https://github.com/thomaspoignant/go-feature-flag',
    sponsor: 'https://github.com/sponsors/thomaspoignant',
    openfeature: 'https://openfeature.dev',
    mailchimpURL:
      '//gofeatureflag.us14.list-manage.com/subscribe/post?u=86acc1a78e371bf66a9683672&amp;id=f42abfec51&amp',
    swaggerURL:
      'https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/cmd/relayproxy/docs/swagger.yaml',
    playgroundEvaluationApi:
      'https://editor.api.gofeatureflag.org/v1/feature/evaluate',
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
      /** @type {import("@docusaurus/preset-classic").Options} */
      ({
        googleAnalytics: {
          trackingID: 'G-LEJBB94YBE',
        },
        gtag: {
          trackingID: 'G-LEJBB94YBE',
        },
        googleTagManager: {
          containerId: 'GTM-MLZMZ3VT',
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
          customCss: [require.resolve('./src/css/custom.css')],
        },
        sitemap: {
          changefreq: 'weekly',
          priority: 0.5,
          ignorePatterns: ['/tags/**'],
          filename: 'sitemap.xml',
        },
      }),
    ],
  ],
  stylesheets: [
    'https://fonts.googleapis.com/css?family=Poppins:400,500,700',
    'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css',
    'https://cdn.jsdelivr.net/gh/devicons/devicon@v2.16.0/devicon.min.css', // https://devicon.dev/
  ],
  themes: [['@docusaurus/theme-mermaid', {theme: 'default'}]],
  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      announcementBar: {
        id: 'support_usz', // Increment on change
        content: `⭐ If you like GO Feature Flag, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/thomaspoignant/go-feature-flag">GitHub</a> and follow us on <a target="_blank" rel="noopener noreferrer" href="https://bsky.app/profile/gofeatureflag.org">Bluesky</a>`,
      },
      image: 'img/logo/x-card.png',
      navbar: {
        title: 'GO Feature Flag',
        logo: {
          alt: 'GO Feature Flag Logo',
          src: 'img/logo/navbar.png',
        },
        items: [
          {
            label: 'SDKs',
            type: 'dropdown',
            className: 'dyte-dropdown',
            items: [
              {
                type: 'html',
                value: generateSdksDropdownHTML(),
                className: 'dyte-dropdown',
              },
            ],
          },
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
              },
            ],
          },
          {
            position: 'left',
            label: 'Developers',
            items: [
              {
                to: '/docs/getting-started',
                html: '<i class="fa-solid fa-rocket menu-icon"></i> Getting Started',
              },
              {
                to: '/docs/sdk',
                html: '<i class="fa-solid fa-code menu-icon"></i> SDKs',
              },
              {
                to: '/editor',
                html: '<i class="fa-solid fa-pencil menu-icon"></i> Flag Editor',
              },
              {
                html: '<i class="fa-solid fa-book menu-icon"></i> Documentation',
                type: 'doc',
                docId: 'index',
              },
              {
                html: '<i class="fa-solid fa-eye menu-icon"></i> Examples <i class="fa fa-external-link" aria-hidden="true"></i>',
                to: 'https://github.com/thomaspoignant/go-feature-flag/tree/main/examples',
              },
              {
                html: '<i class="fa-solid fa-star menu-icon"></i> Feature Flag Best Practice',
                to: '/blog/feature-flag-best-practice',
              },
              {
                to: '/slack',
                html: '<i class="fa-brands fa-slack menu-icon"></i> Community <i class="fa fa-external-link" aria-hidden="true"></i>',
              },
              {
                to: 'https://github.com/thomaspoignant/go-feature-flag/releases',
                html: '<i class="fa-brands fa-github menu-icon"></i> Changelog <i class="fa fa-external-link" aria-hidden="true"></i>',
              },
            ],
          },
          {type: 'doc', docId: 'index', position: 'left', html: 'Docs'},
          {to: '/blog', label: 'Blog', position: 'left'},
          {to: '/editor', html: 'Editor', position: 'left'},
          {to: '/pricing', html: 'Pricing', position: 'left'},
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
            href: '/slack',
            position: 'right',
            className: 'header-slack-link navbar__right',
            'aria-label': 'Slack',
          },
        ],
      },
      footer: {
        logo: {
          alt: 'GO Feature Flag logo',
          src: 'img/logo/logo_footer.png',
          href: 'https://gofeatureflag.org',
          width: 220,
        },
        links: [
          {
            title: 'Product',
            items: [
              {
                label: 'Getting Started',
                to: '/docs/getting-started',
              },
              {
                label: 'OpenFeature',
                to: '/product/open_feature_support',
              },
              {
                label: 'Documentation',
                to: '/docs',
              },
              {
                label: 'SDKs',
                to: '/docs/sdk',
              },
              {
                label: 'Blog',
                to: '/blog',
              },
              {
                label: 'Flag Editor',
                to: '/editor',
              },
              {
                label: 'Pricing',
                to: '/pricing',
              },
            ],
          },
          {
            title: 'SDKs',
            items: (function () {
              return sdk.map(sdk => {
                return {
                  html: `<a href="/docs/sdk/${sdk.docLink}">${sdk.name}</a>`,
                };
              });
            })(),
          },
          {
            title: 'Community',
            items: [
              {
                html: `
                <a href="/slack" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-slack"></i> Slack&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://x.com/gofeatureflag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-x-twitter"></i> X&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://bsky.app/profile/gofeatureflag.org" target="_blank" rel="noreferrer noopener">
                   <i class="fa-brands fa-bluesky"></i> Bluesky&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://youtube.com/@gofeatureflag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-youtube"></i> Youtube &nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
              {
                html: `
                <a href="https://github.com/thomaspoignant/go-feature-flag" target="_blank" rel="noreferrer noopener">
                  <i class="fa-brands fa-github"></i> Github&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
              {
                html: `
                <a href="mailto:contact@gofeatureflag.org" target="_blank" rel="noreferrer noopener">
                  <i class="fa-regular fa-envelope"></i> Email&nbsp;&nbsp;<i class="fa-solid fa-arrow-up-right-from-square text-xs"></i>
                </a>`,
              },
            ],
          },
        ],
        copyright: `Copyright © 2020-${new Date().getFullYear()} GO Feature Flag.<br/>Build with Docusaurus, <a href="https://www.netlify.com" target="_blank" rel="noreferrer noopener">Powered by Netlify</a>`,
      },
      prism: {
        theme: require('prism-react-renderer').themes.vsLight,
        darkTheme: require('prism-react-renderer').themes.vsDark,
        additionalLanguages: [
          'json',
          'java',
          'scala',
          'toml',
          'php',
          'go',
          'csharp',
          'yaml',
          'python',
          'ruby',
        ],
      },
      colorMode: {
        defaultMode: 'dark',
      },
      algolia: {
        appId: 'OV23HUCYBM',
        apiKey: '37574755624276e0d875552f6bcc2b40',
        indexName: 'goff',
        contextualSearch: true,
      },
    }),
};

module.exports = config;

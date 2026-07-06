// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const fs = require('fs');
const path = require('path');

// Versions still built & served (small rolling window, kept for build speed).
const builtVersions = require('./versions.json');

// Versions that once shipped docs but are no longer built. Their /docs/vX.Y.Z/…
// URLs would 404; we 200-redirect each page to its current-docs equivalent
// (see createRedirects in the client-redirects plugin below).
const removedVersions = fs
  .readdirSync(path.join(__dirname, 'versioned_docs'))
  .filter(name => name.startsWith('version-'))
  .map(name => name.replace(/^version-/, ''))
  .filter(version => !builtVersions.includes(version));

const {sdk} = require('./data/sdk');
const {generateSdksDropdownHTML} = require('./src/components/navbar/sdks');
const {
  generateProductDropdownHTML,
} = require('./src/components/navbar/product');
const {
  generateResourcesDropdownHTML,
} = require('./src/components/navbar/resources');
const {
  generateDevelopersDropdownHTML,
} = require('./src/components/navbar/developers');

/** @type {import("@docusaurus/types").Config} */
const config = {
  title: 'GO Feature Flag',
  tagline: 'Open-source feature flags — built on OpenFeature.',
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
            from: '/product/open_feature_support',
            to: '/product/open-feature',
          },
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
          // --- Removed-version documentation redirects (renamed/moved slugs) ---
          // createRedirects (below) auto-covers removed-version URLs whose path
          // is UNCHANGED in the current docs. The entries here handle old URLs
          // whose slug CHANGED between versions (e.g. underscore -> hyphen), so
          // they don't match a current path and createRedirects can't generate
          // them. Every `to` must be a real current docs path.
          // TODO: extend this list from the Google Search Console
          // "Not found (404)" report.
          {from: '/docs/v1.30.0/getting_started', to: '/docs/getting-started'},
          {from: '/docs/v1.0.0/getting_started', to: '/docs/getting-started'},
          {from: '/docs/v0.28.2/getting_started', to: '/docs/getting-started'},
          {from: '/docs/v0.28.1/getting_started', to: '/docs/getting-started'},
        ],
        // For every current-docs page, emit a redirect from the same path under
        // each removed version, so old deep links (e.g.
        // /docs/v1.30.0/sdk/client_providers/openfeature_javascript) 200-redirect
        // to the current page instead of 404-ing. Only the current version is
        // served at /docs/<path> with no version segment; built versioned paths
        // (/docs/v1.54.1/…) and the /docs/next tree are skipped. Old URLs whose
        // slug changed between versions won't match a current path and are
        // handled by the 404 page (src/theme/NotFound/Content) plus the explicit
        // entries above.
        /** @param {string} existingPath */
        createRedirects(existingPath) {
          // Bare docs landing -> emit /docs/<version> for old bare version roots.
          if (existingPath === '/docs') {
            return removedVersions.map(version => `/docs/${version}`);
          }
          if (!existingPath.startsWith('/docs/')) {
            return undefined;
          }
          const subPath = existingPath.slice('/docs/'.length);
          const firstSegment = subPath.split('/')[0];
          if (
            firstSegment === 'next' ||
            /^v\d+\.\d+\.\d+(?:-rc\.\d+)?$/.test(firstSegment)
          ) {
            return undefined;
          }
          return removedVersions.map(version => `/docs/${version}/${subPath}`);
        },
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
            type: 'dropdown',
            className: 'dyte-dropdown',
            items: [
              {
                type: 'html',
                value: generateProductDropdownHTML(),
                className: 'dyte-dropdown',
              },
            ],
          },
          {
            position: 'left',
            label: 'Resources',
            type: 'dropdown',
            className: 'dyte-dropdown',
            items: [
              {
                type: 'html',
                value: generateResourcesDropdownHTML(),
                className: 'dyte-dropdown',
              },
            ],
          },
          {
            position: 'left',
            label: 'Developers',
            type: 'dropdown',
            className: 'dyte-dropdown',
            items: [
              {
                type: 'html',
                value: generateDevelopersDropdownHTML(),
                className: 'dyte-dropdown',
              },
            ],
          },
          {to: '/pricing', html: 'Pricing', position: 'left'},
          {
            to: 'https://github.com/sponsors/thomaspoignant',
            label: 'Sponsor us ❤️',
            position: 'right',
          },
          {
            type: 'custom-githubStars',
            position: 'right',
            className: 'navbar__right',
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
                to: '/product/open-feature',
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

import React from 'react';
import Head from '@docusaurus/Head';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

// Homepage JSON-LD structured data. Emits Organization + SoftwareApplication
// schema so LLMs, search engines, and knowledge graphs can reconcile GO Feature
// Flag as an entity (name, MIT license, free offer, supported languages, links).
// Mirrors the per-page pattern in components/landingPage/seo.
export default function HomeJsonLd() {
  const {siteConfig} = useDocusaurusContext();
  const siteUrl = siteConfig.url;

  const structuredData = [
    {
      '@context': 'https://schema.org',
      '@type': 'Organization',
      name: 'GO Feature Flag',
      url: siteUrl,
      logo: `${siteUrl}/img/logo/logo.png`,
      description: siteConfig.customFields.description,
      sameAs: [
        siteConfig.customFields.github,
        'https://bsky.app/profile/gofeatureflag.org',
        'https://openfeature.dev',
      ],
    },
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: 'GO Feature Flag',
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'Linux, macOS, Windows, Docker, Kubernetes',
      url: siteUrl,
      downloadUrl: siteConfig.customFields.github,
      description:
        'Self-hosted, OpenFeature-native, 100% open-source feature flag ' +
        'solution — no database to operate and no per-seat bill.',
      license: 'https://opensource.org/licenses/MIT',
      // Supported SDK languages — carries the multi-language wedge.
      programmingLanguage: [
        'Go',
        'Java',
        'Kotlin',
        'JavaScript',
        'TypeScript',
        'Python',
        '.NET',
        'Ruby',
        'Swift',
        'PHP',
        'Rust',
      ],
      author: {'@type': 'Person', name: 'Thomas Poignant'},
      offers: {'@type': 'Offer', price: '0', priceCurrency: 'USD'},
    },
  ];

  return (
    <Head>
      <script type="application/ld+json">
        {JSON.stringify(structuredData)}
      </script>
    </Head>
  );
}

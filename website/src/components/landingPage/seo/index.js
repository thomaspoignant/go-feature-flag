import React from 'react';
import PropTypes from 'prop-types';
import Head from '@docusaurus/Head';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

// Shared <head> block for the /product/* landing pages: emits the canonical
// link, the JSON-LD structured data (TechArticle + breadcrumb) and, when an
// image is provided, the Open Graph / Twitter social-card meta tags. Keeping it
// in one place avoids repeating the same SEO boilerplate on every page.
export default function SeoHead({
  title,
  description,
  path,
  image,
  imageMeta,
  imageWidth,
  imageHeight,
  datePublished,
  dateModified,
}) {
  const {siteConfig} = useDocusaurusContext();
  const siteUrl = siteConfig.url;
  const pageUrl = `${siteUrl}${path}`;
  const imageUrl = image ? `${siteUrl}${image}` : undefined;

  const structuredData = [
    {
      '@context': 'https://schema.org',
      '@type': 'TechArticle',
      headline: title,
      description,
      ...(imageUrl ? {image: imageUrl} : {}),
      author: {'@type': 'Organization', name: siteConfig.title, url: siteUrl},
      publisher: {
        '@type': 'Organization',
        name: siteConfig.title,
        logo: {'@type': 'ImageObject', url: `${siteUrl}/img/logo/logo.png`},
      },
      datePublished,
      dateModified,
      mainEntityOfPage: {'@type': 'WebPage', '@id': pageUrl},
    },
    {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: [
        {'@type': 'ListItem', position: 1, name: 'Home', item: `${siteUrl}/`},
        {'@type': 'ListItem', position: 2, name: title, item: pageUrl},
      ],
    },
  ];

  return (
    <Head>
      <link rel="canonical" href={pageUrl} />
      {imageMeta && imageUrl && <meta property="og:image" content={imageUrl} />}
      {imageMeta && imageUrl && imageWidth && (
        <meta property="og:image:width" content={String(imageWidth)} />
      )}
      {imageMeta && imageUrl && imageHeight && (
        <meta property="og:image:height" content={String(imageHeight)} />
      )}
      {imageMeta && imageUrl && (
        <meta name="twitter:image" content={imageUrl} />
      )}
      <script type="application/ld+json">
        {JSON.stringify(structuredData)}
      </script>
    </Head>
  );
}

SeoHead.propTypes = {
  // Page title, reused as the JSON-LD headline and breadcrumb leaf.
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
  // Absolute path on the site, e.g. '/product/integrations'.
  path: PropTypes.string.isRequired,
  // Social-card image path relative to the site URL, e.g. '/img/logo/x-card.png'.
  image: PropTypes.string,
  // When true, also emit og:image / twitter:image meta tags for `image`.
  imageMeta: PropTypes.bool,
  imageWidth: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  imageHeight: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  datePublished: PropTypes.string,
  dateModified: PropTypes.string,
};

SeoHead.defaultProps = {
  image: undefined,
  imageMeta: false,
  imageWidth: undefined,
  imageHeight: undefined,
  datePublished: '2026-06-30',
  dateModified: '2026-06-30',
};

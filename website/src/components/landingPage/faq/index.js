import React from 'react';
import PropTypes from 'prop-types';
import Head from '@docusaurus/Head';

// Flatten a React node into plain text for the FAQ JSON-LD payload.
function toPlainText(node) {
  if (node == null || node === false || node === true) return '';
  if (typeof node === 'string' || typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(toPlainText).join('');
  if (node.props?.children) return toPlainText(node.props.children);
  return '';
}

export default function Faq({title, items, withJsonLd}) {
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: items.map(item => ({
      '@type': 'Question',
      name: toPlainText(item.question),
      acceptedAnswer: {'@type': 'Answer', text: toPlainText(item.answer)},
    })),
  };

  return (
    <section className="px-6 py-12 lg:px-16 xl:px-64 2xl:px-96">
      {withJsonLd && (
        <Head>
          <script type="application/ld+json">{JSON.stringify(jsonLd)}</script>
        </Head>
      )}
      {title && (
        <h2 className="mb-8 text-center text-3xl font-bold text-gray-800 dark:text-gray-50">
          {title}
        </h2>
      )}
      <div className="flex flex-col gap-3">
        {items.map(item => (
          <details
            key={toPlainText(item.question)}
            className="group rounded-xl border border-solid border-gray-200 bg-white dark:border-gray-700 dark:bg-[#1f2024]">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-4 p-5 text-lg font-semibold text-gray-800 [&::-webkit-details-marker]:hidden dark:text-gray-50">
              {item.question}
              <span
                aria-hidden="true"
                className="shrink-0 text-2xl leading-none text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] transition-transform duration-200 group-open:rotate-45">
                +
              </span>
            </summary>
            <div className="px-5 pb-5 leading-relaxed text-gray-700 dark:text-gray-300">
              {item.answer}
            </div>
          </details>
        ))}
      </div>
    </section>
  );
}

Faq.propTypes = {
  title: PropTypes.node,
  items: PropTypes.arrayOf(
    PropTypes.shape({
      question: PropTypes.node.isRequired,
      answer: PropTypes.node.isRequired,
    })
  ).isRequired,
  // Emit FAQPage JSON-LD structured data for SEO.
  withJsonLd: PropTypes.bool,
};

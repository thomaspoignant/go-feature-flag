import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

function Media({
  imageSrc,
  imageAlt,
  placeholderLabel,
  imageClassName,
  imageWidth,
  imageHeight,
}) {
  if (imageSrc) {
    return (
      <img
        src={imageSrc}
        alt={imageAlt}
        width={imageWidth}
        height={imageHeight}
        loading="lazy"
        decoding="async"
        className={
          imageClassName ??
          'mx-auto h-auto w-full max-w-xl rounded-2xl md:max-w-none'
        }
      />
    );
  }
  return (
    <div className="mx-auto flex aspect-[4/3] w-full max-w-xl flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed border-gray-300 bg-gray-50 p-6 text-center dark:border-gray-600 dark:bg-[#1f2024] md:max-w-none">
      <span className="text-3xl" aria-hidden="true">
        🖼️
      </span>
      <span className="text-sm font-semibold text-gray-500 dark:text-gray-400">
        Illustration placeholder
      </span>
      {placeholderLabel && (
        <span className="text-xs text-gray-400">{placeholderLabel}</span>
      )}
    </div>
  );
}

Media.propTypes = {
  imageSrc: PropTypes.string,
  imageAlt: PropTypes.string,
  placeholderLabel: PropTypes.node,
  imageClassName: PropTypes.string,
  imageWidth: PropTypes.number,
  imageHeight: PropTypes.number,
};

export default function FeatureRow({
  eyebrow,
  title,
  children,
  imageSrc,
  imageAlt,
  imageClassName,
  imageWidth,
  imageHeight,
  placeholderLabel,
  reverse,
  actions,
}) {
  return (
    <section className="px-6 py-2 lg:px-16 xl:px-64">
      <div
        className={`flex flex-col items-center gap-10 lg:gap-16 ${
          reverse ? 'md:flex-row-reverse' : 'md:flex-row'
        }`}>
        <div className="md:flex-1">
          {eyebrow && (
            <span className="mb-2 block text-sm font-semibold uppercase tracking-wide text-titles-500">
              {eyebrow}
            </span>
          )}
          {title && (
            <h2 className="mb-4 mt-0 text-3xl font-bold text-gray-800 dark:text-gray-50 sm:text-4xl">
              {title}
            </h2>
          )}
          <div className="text-lg leading-relaxed text-gray-700 dark:text-gray-300">
            {children}
          </div>
          {actions && actions.length > 0 && (
            <div className="mt-6 flex flex-wrap gap-4">
              {actions.map(action => (
                <Link
                  key={action.href}
                  to={action.href}
                  className="inline-flex items-center gap-1 font-semibold text-[color:var(--ifm-color-primary-dark)] dark:text-[color:var(--ifm-color-primary)] no-underline hover:no-underline hover:opacity-80">
                  {action.label} <span aria-hidden="true">→</span>
                </Link>
              ))}
            </div>
          )}
        </div>
        <div className="w-full md:flex-1">
          <Media
            imageSrc={imageSrc}
            imageAlt={imageAlt}
            imageClassName={imageClassName}
            imageWidth={imageWidth}
            imageHeight={imageHeight}
            placeholderLabel={placeholderLabel}
          />
        </div>
      </div>
    </section>
  );
}

FeatureRow.propTypes = {
  eyebrow: PropTypes.node,
  title: PropTypes.node,
  children: PropTypes.node,
  // When omitted, a styled placeholder box is rendered instead of an <img>.
  imageSrc: PropTypes.string,
  imageAlt: PropTypes.string,
  imageClassName: PropTypes.string,
  // Intrinsic pixel dimensions — set both to reserve space and avoid layout shift.
  imageWidth: PropTypes.number,
  imageHeight: PropTypes.number,
  placeholderLabel: PropTypes.node,
  // Put the media on the left (text on the right) when true.
  reverse: PropTypes.bool,
  // Optional inline text links rendered under the body.
  actions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.node.isRequired,
      href: PropTypes.string.isRequired,
    })
  ),
};

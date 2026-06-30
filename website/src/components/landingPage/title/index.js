import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

// Decorative teal/green blobs, mirroring the homepage hero background.
function HeroShapes() {
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 z-0 overflow-hidden">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 898.595 671.33"
        className="absolute right-0 top-0 w-2/5 max-w-md">
        <g>
          <path
            d="M77.225.84S65.754 56.25 72.99 108.615c6.519 47.174 14.071 83.313 45.359 132.19s67.663 74.2 113.344 90.467c10.087 2.544 22.468 4.651 35.375 10.446 17.912 8.956 39.851 18.784 63.185 64.959 29.724 58.823 31.289 129.222 102.94 193.364 30.523 27.324 85.56 48.346 152.718 60.609 90.568 16.538 203.528 16.044 311.709-17.053 0-46.584-.016-642.734-.016-642.734z"
            fill="#cdf7e7"
          />
          <path
            d="M4.946.863s-11.953 71.78 4.545 135.746 50.072 127.642 106.223 162.953c30.391 18.524 54.077 22.62 54.077 22.62s35.965 6.587 58.362 28.851 33.95 47.335 40.287 63.6 30.656 87.859 39.048 101.217 22.093 51.037 70.9 84.776 130.668 56.964 257.731 60.438 240.329-44.6 261.458-55.888"
            fill="none"
            stroke="#273437"
            strokeLinecap="round"
            strokeWidth="1.5"
          />
        </g>
      </svg>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 266.025 234.723"
        className="absolute left-0 top-0 w-[12%] max-w-[120px]">
        <g>
          <path
            d="M246.353.908s32.245 23.839 14.178 60.475-62.607 44.54-85.191 84.439-37.268 86.821-83.942 85.186c-36.9-1.289-90.335-44.54-90.335-44.54L1.188.657z"
            fill="#cdf7e7"
          />
          <path
            d="M178.777.908s4.869 19.639 32.623 49.99 37.492 50.964 33.6 76.932-38.147 51.938-67.684 70.765-58.263 37.33-101.763 35.22c-40.1-1.945-59.9-27.777-71.327-57.935-.862-2.273-1.961-6.034-3.171-6.986"
            fill="none"
            stroke="#273437"
            strokeLinecap="round"
            strokeWidth="1.5"
          />
        </g>
      </svg>
    </div>
  );
}

function HeroActions({actions}) {
  return (
    <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
      {actions.map(action => {
        const secondary = action.variant === 'secondary';
        return (
          <Link
            key={action.href}
            to={action.href}
            className={
              secondary
                ? 'inline-flex items-center gap-2 rounded-lg border border-solid border-gray-300 bg-transparent px-6 py-3 text-base font-semibold text-gray-800 no-underline hover:border-gray-400 hover:no-underline dark:border-gray-600 dark:text-gray-100'
                : 'inline-flex items-center gap-2 rounded-lg bg-[color:var(--ifm-color-primary)] px-6 py-3 text-base font-semibold text-gray-600 no-underline hover:text-gray-800 hover:no-underline hover:brightness-110'
            }>
            {action.icon}
            {action.label}
          </Link>
        );
      })}
    </div>
  );
}

HeroActions.propTypes = {
  actions: PropTypes.array.isRequired,
};

export default function Title({title, description, actions}) {
  return (
    <section className="relative overflow-hidden px-6 py-16 text-center sm:py-20">
      <HeroShapes />
      <div className="relative z-10">
        <h1 className="m-0 text-5xl font-extrabold leading-[1.05] tracking-[-0.08rem] text-gray-800 dark:text-gray-50 sm:text-6xl md:text-7xl md:tracking-[-0.12rem]">
          {title}
        </h1>
        <p className="mx-auto mt-5 max-w-2xl text-lg leading-relaxed text-[color:var(--goff-main-ff-description)]">
          {description}
        </p>
        {actions && actions.length > 0 && <HeroActions actions={actions} />}
      </div>
    </section>
  );
}

Title.propTypes = {
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
  // Optional CTA buttons: [{label, href, icon?, variant?: 'primary'|'secondary'}]
  actions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.node.isRequired,
      href: PropTypes.string.isRequired,
      icon: PropTypes.node,
      variant: PropTypes.oneOf(['primary', 'secondary']),
    })
  ),
};

import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

export default function CtaBand({title, description, actions}) {
  return (
    <section className="px-6 py-12 lg:px-16 xl:px-64 2xl:px-96">
      <div className="relative overflow-hidden rounded-2xl bg-[#9fbeb3] px-6 py-14 text-center shadow-lg">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 opacity-30"
          style={{
            backgroundImage:
              'radial-gradient(rgba(4,42,38,0.25) 1.5px, transparent 1.5px)',
            backgroundSize: '18px 18px',
          }}
        />
        <h2 className="relative m-0 text-3xl font-extrabold text-goff-950 sm:text-4xl">
          {title}
        </h2>
        {description && (
          <p className="relative mx-auto mt-4 max-w-2xl text-lg text-goff-900">
            {description}
          </p>
        )}
        {actions && actions.length > 0 && (
          <div className="relative mt-8 flex flex-wrap items-center justify-center gap-3">
            {actions.map(action => {
              const secondary = action.variant === 'secondary';
              return (
                <Link
                  key={action.href}
                  to={action.href}
                  className={
                    secondary
                      ? 'inline-flex items-center gap-2 rounded-lg border border-solid border-goff-900/50 bg-transparent px-6 py-3 text-base font-semibold text-goff-900 no-underline hover:bg-goff-900/10 hover:text-goff-900 hover:no-underline'
                      : 'inline-flex items-center gap-2 rounded-lg bg-goff-900 px-6 py-3 text-base font-semibold text-white no-underline hover:bg-goff-800 hover:text-white hover:no-underline'
                  }>
                  {action.icon}
                  {action.label}
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
}

CtaBand.propTypes = {
  title: PropTypes.node.isRequired,
  description: PropTypes.node,
  actions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.node.isRequired,
      href: PropTypes.string.isRequired,
      icon: PropTypes.node,
      variant: PropTypes.oneOf(['primary', 'secondary']),
    })
  ),
};

import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';

export default function CtaBand({title, description, actions}) {
  return (
    <section className="px-6 py-12 lg:px-16 xl:px-64 2xl:px-96">
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-goff-800 via-goff-700 to-goff-500 px-6 py-14 text-center shadow-lg">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 opacity-20"
          style={{
            backgroundImage:
              'radial-gradient(rgba(255,255,255,0.9) 1.5px, transparent 1.5px)',
            backgroundSize: '18px 18px',
          }}
        />
        <h2 className="relative m-0 text-3xl font-extrabold text-white sm:text-4xl">
          {title}
        </h2>
        {description && (
          <p className="relative mx-auto mt-4 max-w-2xl text-lg text-goff-50">
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
                      ? 'inline-flex items-center gap-2 rounded-lg border border-solid border-white/70 bg-transparent px-6 py-3 text-base font-semibold text-white no-underline hover:bg-white/10 hover:text-white hover:no-underline'
                      : 'inline-flex items-center gap-2 rounded-lg bg-white px-6 py-3 text-base font-semibold text-goff-800 no-underline hover:bg-goff-50 hover:no-underline'
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

import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';

const COLUMNS = [
  {
    key: 'goff',
    name: 'GO Feature Flag',
    tagline: 'Open source · file-based',
    featured: true,
  },
  {
    key: 'saas',
    name: 'SaaS platforms',
    tagline: 'LaunchDarkly · Split ...',
  },
  {
    key: 'db',
    name: 'DB-backed OSS',
    tagline: 'Unleash · Flagsmith ...',
  },
  {
    key: 'diy',
    name: 'DIY config',
    tagline: 'Roll your own solution',
  },
];

const ROWS = [
  {
    label: 'OpenFeature Support',
    cells: {
      goff: {
        icon: 'check',
        text: 'Support only OpenFeature (no vendor lock-in)',
      },
      saas: {text: 'limited support'},
      db: {text: 'limited support'},
      diy: {icon: 'cross'},
    },
  },
  {
    label: 'Pricing',
    cells: {
      goff: {text: ' — MIT License', strong: 'Free'},
      saas: {text: 'Per-seat / per-MAU'},
      db: {text: 'Free core, paid tiers'},
      diy: {text: 'Free'},
    },
  },
  {
    label: 'Database required',
    cells: {
      goff: {strong: 'None', text: ' — flags are a file'},
      saas: {text: 'Vendor-managed'},
      db: {text: 'PostgreSQL'},
      diy: {text: 'None'},
    },
  },
  {
    label: 'Self-hosting',
    cells: {
      goff: {icon: 'check', text: 'runs in your infra'},
      saas: {icon: 'cross', text: 'vendor-hosted only', crossText: true},
      db: {icon: 'check', text: 'limited self-hosting options'},
      diy: {icon: 'check'},
    },
  },
  {
    label: 'Rollouts & targeting',
    cells: {
      goff: {icon: 'check'},
      saas: {icon: 'check'},
      db: {icon: 'check'},
      diy: {icon: 'cross', text: 'you build it', crossText: true},
    },
  },
  {
    label: 'A/B testing',
    cells: {
      goff: {icon: 'check'},
      saas: {icon: 'check'},
      db: {text: 'Varies'},
      diy: {icon: 'cross'},
    },
  },
  {
    label: 'Collect Usage data',
    cells: {
      goff: {icon: 'check', text: 'via exporters'},
      saas: {icon: 'check'},
      db: {text: 'Partial / paid'},
      diy: {icon: 'cross', text: 'you build it', crossText: true},
    },
  },
];

const borderClass = 'border-[#273437]/10 dark:border-white/10';

CellValue.propTypes = {
  data: PropTypes.shape({
    icon: PropTypes.oneOf(['check', 'cross']),
    strong: PropTypes.string,
    text: PropTypes.string,
    crossText: PropTypes.bool,
  }),
};

function CellValue({data}) {
  if (!data) {
    return (
      <span className="inline-flex items-center justify-center flex-wrap gap-1.5 text-[0.82rem] leading-snug text-gray-800 dark:text-gray-50" />
    );
  }
  const {icon, strong, text, crossText} = data;
  const stack = (icon === 'check' || icon === 'cross') && (strong || text);
  return (
    <span
      className={clsx(
        'inline-flex items-center justify-center text-[0.82rem] leading-snug text-gray-800 dark:text-gray-50',
        stack ? 'flex-col gap-1' : 'flex-wrap gap-1.5'
      )}>
      {icon === 'check' && (
        <span
          className="inline-flex items-center justify-center w-5 h-5 rounded-full shrink-0 bg-[#73c6b6] text-white"
          aria-hidden>
          <i className="fa-solid fa-check" />
        </span>
      )}
      {icon === 'cross' && (
        <span
          className="inline-flex items-center justify-center w-5 h-5 rounded-full shrink-0 bg-red-400/15 text-red-700 dark:text-red-400"
          aria-hidden>
          <i className="fa-solid fa-xmark" />
        </span>
      )}
      {strong && <strong>{strong}</strong>}
      {text && (
        <span className={crossText ? 'text-gray-500 dark:text-gray-400' : ''}>
          {text}
        </span>
      )}
    </span>
  );
}

export function WhyGoff() {
  const lastRowIndex = ROWS.length - 1;
  return (
    <section
      className="py-12 px-4 pb-16 bg-gray-100 dark:bg-[#363636] lg:px-6"
      aria-labelledby="why-goff-title">
      <div className="max-w-6xl mx-auto">
        <h2 id="why-goff-title" className="goffMainTitle text-center">
          Why GO Feature Flag?
        </h2>
        <p className="mt-3 mb-10 max-w-2xl mx-auto text-center text-gray-500 dark:text-gray-400 text-[1.05rem] leading-relaxed">
          Compare file-based open source against SaaS platforms, database-backed
          tools, and rolling your own.
        </p>

        <div className="overflow-x-auto pb-2">
          <div
            className={clsx(
              'grid min-w-[52rem] rounded-2xl overflow-hidden bg-white dark:bg-[#2a2a2a] border',
              borderClass,
              'grid-cols-[minmax(9rem,1.15fr)_repeat(4,minmax(8.5rem,1fr))]'
            )}
            role="table"
            aria-label="Feature flag solution comparison">
            <div className="contents" role="row">
              <div
                className="min-h-[7.5rem] bg-titles-500/10"
                role="columnheader"
                aria-hidden
              />

              {COLUMNS.map(column => (
                <div
                  key={column.key}
                  className={clsx(
                    'relative flex flex-col justify-center items-center min-h-[8.5rem] px-3 py-4 text-center',
                    column.featured &&
                      clsx(
                        'border-solid border-b',
                        borderClass,
                        'bg-titles-500/25 border-t-2 border-l-2 border-r-2 !border-t-titles-500 !border-l-titles-500 !border-r-titles-500 rounded-t-xl'
                      )
                  )}
                  role="columnheader">
                  {column.badge && (
                    <span className="absolute top-2.5 right-2.5 px-2 py-0.5 rounded-full bg-titles-500 text-white text-[0.68rem] font-bold tracking-wide uppercase">
                      {column.badge}
                    </span>
                  )}
                  <h3 className="m-0 text-[1.22rem] font-bold leading-snug text-gray-800 dark:text-gray-50">
                    {column.name}
                  </h3>
                  <p className="mt-1.5 mb-0 text-xs leading-snug text-gray-500 dark:text-gray-400">
                    {column.tagline}
                  </p>
                </div>
              ))}
            </div>

            {ROWS.map((row, rowIndex) => {
              const isLastRow = rowIndex === lastRowIndex;
              const isAltRow = rowIndex % 2 === 1;

              return (
                <div className="contents" role="row" key={row.label}>
                  <div
                    className={clsx(
                      'flex items-center min-h-[3.25rem] px-4 py-3 text-sm font-semibold text-gray-800 dark:text-gray-50',
                      isAltRow && 'bg-titles-500/10'
                    )}
                    role="rowheader">
                    {row.label}
                  </div>

                  {COLUMNS.map(column => (
                    <div
                      key={`${row.label}-${column.key}`}
                      className={clsx(
                        'flex items-center justify-center min-h-[3.25rem] px-3 py-2.5 text-center',
                        isAltRow && 'bg-titles-500/10',
                        column.featured && isAltRow && 'bg-titles-500/20',
                        column.featured &&
                          clsx(
                            'border-solid',
                            borderClass,
                            'border-l-2 border-r-2 !border-l-titles-500 !border-r-titles-500',
                            isLastRow
                              ? 'border-b-2 !border-b-titles-500 rounded-b-xl'
                              : 'border-b'
                          )
                      )}
                      role="cell">
                      <CellValue data={row.cells[column.key]} />
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}

import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import goffLogo from '@site/static/img/logo/logo.png';
import flagdLogo from '@site/static/img/comparison/flagd.png';
import unleashLogo from '@site/static/img/comparison/unleash.png';
import flagsmithLogo from '@site/static/img/comparison/flagsmith.png';
import growthbookLogo from '@site/static/img/comparison/growthbook.png';
import fliptLogo from '@site/static/img/comparison/flipt.png';
import posthogLogo from '@site/static/img/comparison/posthog.png';
import featurevisorLogo from '@site/static/img/comparison/featurevisor.png';

// Columns mirror the comparison spec. The first column is the row header (tool).
const COLUMNS = [
  {key: 'tool', label: 'Tool'},
  {key: 'license', label: 'License'},
  {key: 'db', label: 'DB required?'},
  {key: 'openfeature', label: 'OpenFeature support'},
  {key: 'ui', label: 'Built-in UI'},
  {key: 'experimentation', label: 'Experimentation / A-B'},
  {key: 'open', label: 'Truly open (no paywalled core)?'},
];

// Every cell value below is verified against the project's own docs / the
// OpenFeature ecosystem (June 2026). Keep this qualitative — no numbers.
// tone: 'good' | 'bad' | 'warn' | null (neutral). icon: 'check' | 'cross' | 'partial'.
const ROWS = [
  {
    featured: true,
    tool: 'GO Feature Flag',
    logo: goffLogo,
    cells: {
      stack: {text: 'Go'},
      license: {text: 'MIT', tone: 'good'},
      db: {text: 'None', icon: 'check', tone: 'good'},
      openfeature: {text: 'Native (ecosystem)', icon: 'check', tone: 'good'},
      ui: {
        text: 'Config editor (no runtime admin UI)',
        tone: 'warn',
        icon: 'partial',
      },
      experimentation: {
        text: 'A/B + data exporters',
        icon: 'check',
        tone: 'good',
      },
      open: {text: 'Yes — fully open', icon: 'check', tone: 'good'},
    },
  },
  {
    tool: 'flagd',
    logo: flagdLogo,
    cells: {
      stack: {text: 'Go'},
      license: {text: 'Apache-2.0', tone: 'good'},
      db: {text: 'None', icon: 'check', tone: 'good'},
      openfeature: {
        text: 'Native (reference impl)',
        icon: 'check',
        tone: 'good',
      },
      ui: {text: 'None (headless daemon)', icon: 'cross'},
      experimentation: {text: 'No', icon: 'cross'},
      open: {text: 'Yes — fully open', icon: 'check', tone: 'good'},
    },
  },
  {
    tool: 'Unleash',
    logo: unleashLogo,
    cells: {
      stack: {text: 'Node.js'},
      license: {text: 'Apache-2.0 (open-core)'},
      db: {text: 'PostgreSQL', icon: 'cross'},
      openfeature: {text: 'Community provider', icon: 'partial', tone: 'warn'},
      ui: {text: 'Yes', icon: 'check'},
      experimentation: {text: 'Partial', icon: 'partial', tone: 'warn'},
      open: {text: 'No — paid Enterprise tier', icon: 'cross', tone: 'bad'},
    },
  },
  {
    tool: 'Flagsmith',
    logo: flagsmithLogo,
    cells: {
      stack: {text: 'Python / Django'},
      license: {text: 'BSD-3 (open-core)'},
      db: {text: 'PostgreSQL', icon: 'cross'},
      openfeature: {text: 'Official provider', icon: 'check', tone: 'good'},
      ui: {text: 'Yes', icon: 'check'},
      experimentation: {text: 'Partial', icon: 'partial', tone: 'warn'},
      open: {text: 'No — paid Enterprise tier', icon: 'cross', tone: 'bad'},
    },
  },
  {
    tool: 'GrowthBook',
    logo: growthbookLogo,
    cells: {
      stack: {text: 'Node.js'},
      license: {text: 'MIT (open-core)'},
      db: {text: 'MongoDB', icon: 'cross'},
      openfeature: {text: 'Official provider', icon: 'check', tone: 'good'},
      ui: {text: 'Yes', icon: 'check'},
      experimentation: {
        text: 'Yes — core strength',
        icon: 'check',
        tone: 'good',
      },
      open: {text: 'No — commercial tiers', icon: 'cross', tone: 'bad'},
    },
  },
  {
    tool: 'Flipt',
    logo: fliptLogo,
    cells: {
      stack: {text: 'Go'},
      license: {text: 'MIT client / FCL server'},
      db: {text: 'None (v2, Git-native)', icon: 'check', tone: 'good'},
      openfeature: {text: 'Provider + OFREP', icon: 'check', tone: 'good'},
      ui: {text: 'Yes', icon: 'check'},
      experimentation: {text: 'Partial', icon: 'partial', tone: 'warn'},
      open: {text: 'Source-available server', icon: 'partial', tone: 'warn'},
    },
  },
  {
    tool: 'PostHog',
    logo: posthogLogo,
    cells: {
      stack: {text: 'Python'},
      license: {text: 'MIT (open-core)'},
      db: {text: 'Postgres + ClickHouse + Redis', icon: 'cross', tone: 'bad'},
      openfeature: {text: 'Community provider', icon: 'partial', tone: 'warn'},
      ui: {text: 'Yes (full suite)', icon: 'check'},
      experimentation: {text: 'Yes', icon: 'check', tone: 'good'},
      open: {text: 'No — enterprise dir', icon: 'cross', tone: 'bad'},
    },
  },
  {
    tool: 'Featurevisor',
    logo: featurevisorLogo,
    cells: {
      stack: {text: 'TypeScript'},
      license: {text: 'MIT', tone: 'good'},
      db: {text: 'None (GitOps)', icon: 'check', tone: 'good'},
      openfeature: {text: 'None (own SDKs)', icon: 'cross', tone: 'bad'},
      ui: {text: 'None (Git/CLI)', icon: 'cross'},
      experimentation: {text: 'Definition only', icon: 'partial', tone: 'warn'},
      open: {text: 'Yes — fully open', icon: 'check', tone: 'good'},
    },
  },
];

const borderClass = 'border-[#273437]/10 dark:border-white/10';

function ToneIcon({icon}) {
  if (icon === 'check') {
    return (
      <span
        className="inline-flex items-center justify-center w-5 h-5 rounded-full shrink-0 bg-[#73c6b6] text-white"
        aria-hidden>
        <i className="fa-solid fa-check" />
      </span>
    );
  }
  if (icon === 'cross') {
    return (
      <span
        className="inline-flex items-center justify-center w-5 h-5 rounded-full shrink-0 bg-red-400/15 text-red-700 dark:text-red-400"
        aria-hidden>
        <i className="fa-solid fa-xmark" />
      </span>
    );
  }
  if (icon === 'partial') {
    return (
      <span
        className="inline-flex items-center justify-center w-5 h-5 rounded-full shrink-0 bg-amber-400/20 text-amber-700 dark:text-amber-400"
        aria-hidden>
        <i className="fa-solid fa-circle-half-stroke" />
      </span>
    );
  }
  return null;
}

ToneIcon.propTypes = {
  icon: PropTypes.oneOf(['check', 'cross', 'partial']),
};

const toneTextClass = {
  good: 'text-gray-800 dark:text-gray-50',
  warn: 'text-amber-700 dark:text-amber-400',
  bad: 'text-red-700 dark:text-red-400',
};

function Cell({data}) {
  if (!data) {
    return null;
  }
  const {text, icon, tone} = data;
  return (
    <span className="inline-flex items-center gap-1.5 text-[0.82rem] leading-snug">
      {icon && <ToneIcon icon={icon} />}
      {text && (
        <span
          className={toneTextClass[tone] || 'text-gray-700 dark:text-gray-200'}>
          {text}
        </span>
      )}
    </span>
  );
}

Cell.propTypes = {
  data: PropTypes.shape({
    text: PropTypes.string,
    icon: PropTypes.oneOf(['check', 'cross', 'partial']),
    tone: PropTypes.oneOf(['good', 'warn', 'bad']),
  }),
};

export function ComparisonTable() {
  return (
    <div className="not-prose overflow-x-auto pb-2 my-8">
      <table
        className={clsx(
          'w-full min-w-[60rem] border-collapse rounded-2xl overflow-hidden bg-white dark:bg-[#2a2a2a] border text-left',
          borderClass
        )}>
        <thead>
          <tr className="bg-titles-500/10">
            {COLUMNS.map(column => (
              <th
                key={column.key}
                scope="col"
                className={clsx(
                  'px-3 py-3 text-[0.8rem] font-bold align-bottom text-gray-800 dark:text-gray-50 border-b',
                  borderClass,
                  column.key === 'tool' && 'sticky left-0 bg-titles-500/10'
                )}>
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {ROWS.map((row, rowIndex) => {
            const isAlt = rowIndex % 2 === 1;
            return (
              <tr
                key={row.tool}
                className={clsx(
                  isAlt && 'bg-titles-500/[0.04]',
                  row.featured && 'bg-titles-500/[0.12]'
                )}>
                <th
                  scope="row"
                  className={clsx(
                    'px-3 py-3 text-[0.92rem] font-bold border-b align-middle whitespace-nowrap sticky left-0',
                    borderClass,
                    row.featured
                      ? 'text-titles-500 bg-titles-500/[0.12]'
                      : clsx(
                          'text-gray-800 dark:text-gray-50',
                          isAlt
                            ? 'bg-titles-500/[0.04]'
                            : 'bg-white dark:bg-[#2a2a2a]'
                        )
                  )}>
                  <span className="inline-flex items-center gap-2">
                    {row.logo && (
                      <img
                        src={row.logo}
                        alt={`${row.tool} logo`}
                        loading="lazy"
                        className="w-6 h-6 shrink-0 rounded object-contain"
                      />
                    )}
                    <span className="inline-flex flex-col items-start gap-1">
                      <span>{row.tool}</span>
                      {row.featured && (
                        <span className="px-2 py-0.5 rounded-full bg-titles-500 text-white text-[0.6rem] font-bold tracking-wide uppercase">
                          Recommended
                        </span>
                      )}
                    </span>
                  </span>
                </th>
                {COLUMNS.slice(1).map(column => (
                  <td
                    key={column.key}
                    className={clsx(
                      'px-3 py-3 align-middle border-b',
                      borderClass
                    )}>
                    <Cell data={row.cells[column.key]} />
                  </td>
                ))}
              </tr>
            );
          })}
        </tbody>
      </table>
      <p className="mt-3 text-xs text-gray-500 dark:text-gray-400">
        <ToneIcon icon="check" /> built-in / fully supported &nbsp;·&nbsp;
        <ToneIcon icon="partial" /> partial or with caveats &nbsp;·&nbsp;
        <ToneIcon icon="cross" /> not available. OpenFeature posture and
        licensing verified against each project&rsquo;s docs and the OpenFeature
        ecosystem (2026). &ldquo;Official&rdquo; = vendor-maintained provider;
        &ldquo;community&rdquo; = third-party provider.
      </p>
    </div>
  );
}

ComparisonTable.propTypes = {};

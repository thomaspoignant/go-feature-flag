import React from 'react';
import PropTypes from 'prop-types';

const BASE_CHIP_CLASS =
  'cursor-pointer rounded-full border border-solid px-4 py-2 text-sm font-semibold transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-goff-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-[#1b1b1d]';

// goff-700 rather than goff-600: white-on-goff-600 is 3.9:1, below the 4.5:1
// WCAG AA threshold for this text size.
const ACTIVE_CHIP_CLASS =
  'border-goff-700 bg-goff-700 text-white dark:border-goff-500 dark:bg-goff-500 dark:text-gray-900';

const INACTIVE_CHIP_CLASS =
  'border-gray-300 bg-transparent text-gray-700 hover:border-goff-500 hover:text-goff-700 dark:border-gray-600 dark:text-gray-200 dark:hover:border-goff-400 dark:hover:text-goff-300';

// `null` means "All articles". Native <button>s give us keyboard support and
// aria-pressed communicates the toggle state to screen readers.
export default function CategoryChips({categories, active, counts, onChange}) {
  const chips = [{value: null, label: 'All articles'}].concat(
    categories.map(category => ({value: category, label: category}))
  );

  return (
    // <fieldset> is the native equivalent of role="group" and groups the chips
    // under the legend for screen readers. Tailwind preflight is disabled on
    // this site, so the browser default border/padding/margin are cleared here,
    // and min-w-0 stops the default `min-inline-size: min-content` from
    // blocking flex-wrap.
    <fieldset className="mx-0 mb-10 flex min-w-0 flex-wrap gap-3 border-0 p-0">
      <legend className="sr-only">Filter articles by category</legend>
      {chips.map(chip => {
        const isActive = chip.value === active;
        const count = chip.value === null ? counts.total : counts[chip.value];
        return (
          <button
            key={chip.label}
            type="button"
            aria-pressed={isActive}
            onClick={() => onChange(chip.value)}
            className={`${BASE_CHIP_CLASS} ${
              isActive ? ACTIVE_CHIP_CLASS : INACTIVE_CHIP_CLASS
            }`}>
            {chip.label}
            <span className="ml-2 opacity-80">{count}</span>
          </button>
        );
      })}
    </fieldset>
  );
}

CategoryChips.propTypes = {
  categories: PropTypes.arrayOf(PropTypes.string).isRequired,
  // The selected category, or null for "All articles".
  active: PropTypes.string,
  // Post count per category, plus `total` for "All articles".
  counts: PropTypes.objectOf(PropTypes.number).isRequired,
  onChange: PropTypes.func.isRequired,
};

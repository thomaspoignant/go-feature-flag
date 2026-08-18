import React from 'react';
import PropTypes from 'prop-types';

export default function CategoryBadge({category}) {
  if (!category) {
    return null;
  }
  return (
    <span className="inline-flex w-fit items-center rounded-full bg-goff-50 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-goff-700 dark:bg-goff-950 dark:text-goff-300">
      {category}
    </span>
  );
}

CategoryBadge.propTypes = {
  category: PropTypes.string,
};

import React from 'react';
import PropTypes from 'prop-types';
import {formatReadingTime} from '../utils';

const authorShape = PropTypes.shape({
  key: PropTypes.string,
  name: PropTypes.string,
  imageURL: PropTypes.string,
});

// Overlapping avatars followed by the author names, then the date and the
// reading time. Rendered inside a card that is itself a link, so nothing in
// here is interactive.
export default function PostMeta({authors, formattedDate, readingTime, size}) {
  const readingTimeLabel = formatReadingTime(readingTime);
  const avatarClass = size === 'large' ? 'h-10 w-10' : 'h-8 w-8';
  const named = authors.filter(author => author.name);

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
      {named.length > 0 && (
        <>
          <div className="flex -space-x-2">
            {named
              .filter(author => author.imageURL)
              .map(author => (
                <img
                  key={author.key ?? author.name}
                  src={author.imageURL}
                  alt=""
                  loading="lazy"
                  className={`${avatarClass} rounded-full border-2 border-solid border-white object-cover dark:border-[#1f2024]`}
                />
              ))}
          </div>
          <span className="text-sm font-semibold text-gray-700 dark:text-gray-200">
            {named.map(author => author.name).join(', ')}
          </span>
        </>
      )}
      <span className="text-sm text-[color:var(--goff-main-ff-description)]">
        {formattedDate}
        {readingTimeLabel ? ` · ${readingTimeLabel}` : ''}
      </span>
    </div>
  );
}

PostMeta.propTypes = {
  authors: PropTypes.arrayOf(authorShape).isRequired,
  formattedDate: PropTypes.string.isRequired,
  readingTime: PropTypes.number,
  // 'large' is used by the hero, everything else by the cards.
  size: PropTypes.oneOf(['default', 'large']),
};

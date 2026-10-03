import PropTypes from 'prop-types';

// Shape produced by `normalizePost` in ./utils.js.
export const postShape = PropTypes.shape({
  permalink: PropTypes.string.isRequired,
  title: PropTypes.string.isRequired,
  description: PropTypes.string,
  formattedDate: PropTypes.string.isRequired,
  readingTime: PropTypes.number,
  authors: PropTypes.array.isRequired,
  image: PropTypes.string.isRequired,
  imageAlt: PropTypes.string.isRequired,
  category: PropTypes.string,
});

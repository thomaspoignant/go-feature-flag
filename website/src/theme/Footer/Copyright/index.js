import React from 'react';
import PropTypes from 'prop-types';

export default function FooterCopyright({copyright}) {
  return (
    <div
      className="footer__copyright"
      // Developer provided the HTML, so assume it's safe.
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={{__html: copyright}}
    />
  );
}

FooterCopyright.propTypes = {
  copyright: PropTypes.string.isRequired,
};

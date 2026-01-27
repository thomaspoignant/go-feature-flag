import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';

export default function FooterLayout({
  style,
  links,
  logo,
  copyright,
  newsletter,
}) {
  return (
    <footer
      className={clsx('footer', {
        'footer--dark': style === 'dark',
      })}>
      <div className={'grid grid-cols-1 md:grid-cols-4'}>
        <div className="max-md:hidden relative overflow-hidden h-full">
          <div className="absolute inset-x-0 top-0 flex justify-left">
            {logo && <div className="margin-bottom--sm">{logo}</div>}
          </div>
          <div className="absolute inset-x-0 bottom-0 p-4 ">{copyright}</div>
        </div>
        <div className={'max-sm:hidden col-span-2'}>{links}</div>
        <div>{newsletter}</div>
      </div>
    </footer>
  );
}

FooterLayout.propTypes = {
  style: PropTypes.string,
  links: PropTypes.node,
  logo: PropTypes.node,
  copyright: PropTypes.node,
  newsletter: PropTypes.node,
};

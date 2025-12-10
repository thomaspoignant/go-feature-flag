import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import LinkItem from '@theme/Footer/LinkItem';
function Separator() {
  return <span className="footer__link-separator">·</span>;
}

Separator.propTypes = {
  separator: PropTypes.string,
};

function SimpleLinkItem({item}) {
  return item.html ? (
    <span
      className={clsx('footer__link-item', item.className)}
      // Developer provided the HTML, so assume it's safe.
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={{__html: item.html}}
    />
  ) : (
    <LinkItem item={item} />
  );
}

SimpleLinkItem.propTypes = {
  item: PropTypes.shape({
    html: PropTypes.string,
    className: PropTypes.string,
    to: PropTypes.string,
    href: PropTypes.string,
    label: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
    prependBaseUrlToHref: PropTypes.bool,
  }).isRequired,
};
export default function FooterLinksSimple({links}) {
  return (
    <div className="footer__links text--center">
      <div className="footer__links">
        {links.map((item, i) => (
          <React.Fragment key={item.label ?? item.href ?? item.to}>
            <SimpleLinkItem item={item} />
            {links.length !== i + 1 && <Separator />}
          </React.Fragment>
        ))}
      </div>
    </div>
  );
}

FooterLinksSimple.propTypes = {
  links: PropTypes.arrayOf(
    PropTypes.shape({
      html: PropTypes.string,
      className: PropTypes.string,
      to: PropTypes.string,
      href: PropTypes.string,
      label: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
      prependBaseUrlToHref: PropTypes.bool,
    })
  ).isRequired,
};

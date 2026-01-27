import React from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import isInternalUrl from '@docusaurus/isInternalUrl';
import IconExternalLink from '@theme/Icon/ExternalLink';

export default function FooterLinkItem({item}) {
  const {to, href, label, prependBaseUrlToHref, className, ...props} = item;
  const toUrl = useBaseUrl(to);
  const normalizedHref = useBaseUrl(href, {forcePrependBaseUrl: true});
  return (
    <Link
      className={clsx('footer__link-item', className)}
      {...(href
        ? {
            href: prependBaseUrlToHref ? normalizedHref : href,
          }
        : {
            to: toUrl,
          })}
      {...props}>
      {label}
      {href && !isInternalUrl(href) && <IconExternalLink />}
    </Link>
  );
}

FooterLinkItem.propTypes = {
  item: PropTypes.shape({
    to: PropTypes.string,
    href: PropTypes.string,
    label: PropTypes.string,
    prependBaseUrlToHref: PropTypes.bool,
    className: PropTypes.string,
  }).isRequired,
};

import React from 'react';
import PropTypes from 'prop-types';
import {isMultiColumnFooterLinks} from '@docusaurus/theme-common';
import FooterLinksMultiColumn from '@theme/Footer/Links/MultiColumn';
import FooterLinksSimple from '@theme/Footer/Links/Simple';

export default function FooterLinks({links}) {
  return isMultiColumnFooterLinks(links) ? (
    <FooterLinksMultiColumn columns={links} />
  ) : (
    <FooterLinksSimple links={links} />
  );
}

FooterLinks.propTypes = {
  links: PropTypes.array.isRequired,
};

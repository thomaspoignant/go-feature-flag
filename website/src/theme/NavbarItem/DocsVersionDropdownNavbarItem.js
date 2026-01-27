import React from 'react';
import PropTypes from 'prop-types';
import {
  useVersions,
  useActiveDocContext,
} from '@docusaurus/plugin-content-docs/client';
import {useDocsPreferredVersion} from '@docusaurus/theme-common';
import {useDocsVersionCandidates} from '@docusaurus/plugin-content-docs/lib/client/docsUtils';
import {translate} from '@docusaurus/Translate';
import {useLocation} from '@docusaurus/router';
import DefaultNavbarItem from '@theme/NavbarItem/DefaultNavbarItem';
import DropdownNavbarItem from '@theme/NavbarItem/DropdownNavbarItem';
import semver from 'semver';

const maxVersionToDisplayPerMajor = 4;
const getVersionMainDoc = version =>
  version.docs.find(doc => doc.id === version.mainDocId);
export default function DocsVersionDropdownNavbarItem({
  mobile,
  docsPluginId,
  dropdownActiveClassDisabled,
  dropdownItemsBefore,
  dropdownItemsAfter,
  ...props
}) {
  const {search, hash} = useLocation();
  const activeDocContext = useActiveDocContext(docsPluginId);
  let versions = useVersions(docsPluginId);

  let versionToKeep = getListVersionsToDisplay(
    versions.filter(item => item.name !== 'current').map(i => i.name)
  );

  versions = versions.filter(item => {
    return versionToKeep.includes(item.name.replace('v', ''));
  });
  const {savePreferredVersionName} = useDocsPreferredVersion(docsPluginId);
  const versionLinks = versions.map(version => {
    // We try to link to the same doc, in another version
    // When not possible, fallback to the "main doc" of the version
    const versionDoc =
      activeDocContext.alternateDocVersions[version.name] ??
      getVersionMainDoc(version);
    return {
      label: version.label,
      // preserve ?search#hash suffix on version switches
      to: `${versionDoc.path}${search}${hash}`,
      isActive: () => version === activeDocContext.activeVersion,
      onClick: () => savePreferredVersionName(version.name),
    };
  });
  const items = [
    ...dropdownItemsBefore,
    ...versionLinks,
    ...dropdownItemsAfter,
  ];
  const dropdownVersion = useDocsVersionCandidates(docsPluginId)[0];
  // Mobile dropdown is handled a bit differently
  const dropdownLabel =
    mobile && items.length > 1
      ? translate({
          id: 'theme.navbar.mobileVersionsDropdown.label',
          message: 'Versions',
          description:
            'The label for the navbar versions dropdown on mobile view',
        })
      : dropdownVersion.label;
  const dropdownTo =
    mobile && items.length > 1
      ? undefined
      : getVersionMainDoc(dropdownVersion).path;
  // We don't want to render a version dropdown with 0 or 1 item. If we build
  // the site with a single docs version (onlyIncludeVersions: ['1.0.0']),
  // We'd rather render a button instead of a dropdown
  if (items.length <= 1) {
    return (
      <DefaultNavbarItem
        {...props}
        mobile={mobile}
        label={dropdownLabel}
        to={dropdownTo}
        isActive={dropdownActiveClassDisabled ? () => false : undefined}
      />
    );
  }
  return (
    <DropdownNavbarItem
      {...props}
      mobile={mobile}
      label={dropdownLabel}
      to={dropdownTo}
      items={items}
      isActive={dropdownActiveClassDisabled ? () => false : undefined}
    />
  );
}

function getListVersionsToDisplay(versionToCheck) {
  const latestMinorVersions = new Map();
  const versionMap = new Map();
  for (const v of versionToCheck.map(semver.parse)) {
    const baseVersion = new semver.SemVer(`${v.major}.${v.minor}.0`).toString();
    if (!versionMap.get(v.major)) {
      versionMap.set(v.major, 0);
    }
    if (!latestMinorVersions.has(baseVersion)) {
      versionMap.set(v.major, versionMap.get(v.major) + 1);
      if (versionMap.get(v.major) <= maxVersionToDisplayPerMajor) {
        latestMinorVersions.set(baseVersion, v);
      }
    }
  }
  return Array.from(latestMinorVersions.values()).map(v => v.toString());
}

DocsVersionDropdownNavbarItem.propTypes = {
  mobile: PropTypes.bool,
  docsPluginId: PropTypes.string,
  dropdownActiveClassDisabled: PropTypes.bool,
  dropdownItemsBefore: PropTypes.array,
  dropdownItemsAfter: PropTypes.array,
};

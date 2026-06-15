import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';
import clsx from 'clsx';
import useGitHubStars from '@site/src/hooks/useGitHubStars';

export default function GitHubStarsNavbarItem() {
  const {siteConfig} = useDocusaurusContext();
  const {stars, failed} = useGitHubStars();

  const ariaLabel = stars
    ? `GitHub repository (${stars} stars)`
    : 'GitHub repository';

  return (
    <Link
      to={siteConfig.customFields.github}
      className="navbar__item inline-flex items-center gap-1.5 font-semibold text-current hover:text-current hover:no-underline hover:opacity-60 transition-opacity"
      aria-label={ariaLabel}
      title={ariaLabel}>
      <i
        className="fa-brands fa-github text-xl leading-none"
        aria-hidden="true"
      />
      {(stars !== null || !failed) && (
        <span className="text-sm tabular-nums inline-flex items-center">
          <i
            className="fa-solid fa-star mr-1 text-[#f5b400]"
            aria-hidden="true"
          />
          <span
            className={clsx(
              'min-w-[4ch] transition-opacity duration-300',
              !stars && 'opacity-0'
            )}>
            {stars ?? '0.0k'}
          </span>
        </span>
      )}
    </Link>
  );
}

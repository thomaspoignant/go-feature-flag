import React, {useEffect, useState} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';
import clsx from 'clsx';

export default function GitHubStarsNavbarItem() {
  const {siteConfig} = useDocusaurusContext();
  const [stars, setStars] = useState(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const shieldsUrl = `https://img.shields.io/github/stars/${siteConfig.organizationName}/${siteConfig.projectName}.json`;

    fetch(shieldsUrl, {signal: controller.signal})
      .then(response => {
        if (!response.ok) {
          throw new Error(`shields.io responded with ${response.status}`);
        }
        return response.json();
      })
      .then(data => {
        const isStarCount =
          typeof data?.message === 'string' &&
          /^[\d.,]+[kmb]?$/i.test(data.message.trim());
        if (isStarCount) {
          setStars(data.message);
        } else {
          setFailed(true);
        }
      })
      .catch(error => {
        if (error.name !== 'AbortError') {
          console.error('Failed to fetch GitHub star count:', error);
          setFailed(true);
        }
      });

    return () => controller.abort();
  }, [siteConfig.organizationName, siteConfig.projectName]);

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
          <i className="fa-solid fa-star mr-1 text-[#f5b400]" aria-hidden="true" />
          <span className={clsx('min-w-[3ch]', !stars && 'opacity-0')}>
            {stars ?? '0.0k'}
          </span>
        </span>
      )}
    </Link>
  );
}

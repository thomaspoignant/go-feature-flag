import {useEffect, useState} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

/**
 * Fetches the GitHub star count from shields.io.
 *
 * Returns `{stars, failed}` where `stars` is the formatted count once available
 * (e.g. `"5.8k"`) and `failed` becomes `true` when the count can't be retrieved.
 * While loading, both are falsy/`null`, letting callers reserve space and only
 * hide the star block on failure.
 */
export default function useGitHubStars() {
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

  return {stars, failed};
}

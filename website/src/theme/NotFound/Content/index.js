import React, {useEffect} from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import Translate from '@docusaurus/Translate';
import Heading from '@theme/Heading';
import Link from '@docusaurus/Link';
import {useAllDocsData} from '@docusaurus/plugin-content-docs/client';
import versions from '@site/versions.json';

// Captures the version token and the trailing sub-path:
//   /docs/v1.50.0/sdk/foo -> ["…", "v1.50.0", "/sdk/foo"]
//   /docs/v1.30.0         -> ["…", "v1.30.0", undefined]
const OLD_VERSION_DOCS_PATH = /^\/docs\/(v\d+\.\d+\.\d+(?:-rc\.\d+)?)(\/.*)?$/;
const stripTrailingSlash = path =>
  path.length > 1 ? path.replace(/\/$/, '') : path;

export default function NotFoundContent({className}) {
  const allDocsData = useAllDocsData();

  // Old doc versions are pruned from versions.json to keep build times down, so
  // their previously-indexed URLs now 404. When we land on the 404 page for one
  // of those removed-version URLs, strip the version segment and send the visitor
  // to the equivalent page in the current docs — falling back to the docs home
  // when that page no longer exists (e.g. the slug was renamed between versions),
  // so we never dead-end on a second 404.
  useEffect(() => {
    const match = window.location.pathname.match(OLD_VERSION_DOCS_PATH);
    // Bail unless this is a versioned docs URL whose version is no longer built.
    if (!match || versions.includes(match[1])) {
      return;
    }

    // Version-strip, e.g. /docs/v1.50.0/sdk/foo -> /docs/sdk/foo
    const target = stripTrailingSlash(`/docs${match[2] || '/'}`);

    // Only deep-link if that page still exists in the current docs.
    const currentDocs = Object.values(allDocsData)
      .flatMap(data => data.versions)
      .find(version => version.isLast);
    const exists = currentDocs?.docs?.some(
      doc => stripTrailingSlash(doc.path) === target
    );

    window.location.replace(
      (exists ? target : '/docs/') +
        window.location.search +
        window.location.hash
    );
  }, [allDocsData]);

  return (
    <main className={clsx('container margin-vert--xl', className)}>
      <div className="row">
        <div className="col col--6 col--offset-3">
          <Heading as="h1" className="hero__title">
            <Translate
              id="theme.NotFound.title"
              description="The title of the 404 page">
              Page Not Found
            </Translate>
          </Heading>
          <p>
            <Translate
              id="theme.NotFound.p1"
              description="The first paragraph of the 404 page">
              We could not find what you were looking for.
            </Translate>
          </p>
          <p>
            The version you are looking for may not be available.
            <br />
            Please check <Link to="/versions">the version page</Link> to find
            the documentation associated to the version you are using.
          </p>
          <p>
            <Translate
              id="theme.NotFound.p2"
              description="The 2nd paragraph of the 404 page">
              Please contact the owner of the site that linked you to the
              original URL and let them know their link is broken.
            </Translate>
          </p>
        </div>
      </div>
    </main>
  );
}

NotFoundContent.propTypes = {
  className: PropTypes.string,
};

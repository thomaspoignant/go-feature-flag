import React, { useEffect, useState } from "react";
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';
import Translate from '@docusaurus/Translate';
import {
  useVersions,
  useLatestVersion,
} from '@docusaurus/plugin-content-docs/client';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

const docsPluginId = undefined; // Default docs plugin instance
const baseGithubLinkDoc = 'https://github.com/thomaspoignant/go-feature-flag/tree/main/website/versioned_docs/version-';

function DocumentationLabel() {
  return (
    <Translate id="versionsPage.versionEntry.link">Documentation</Translate>
  );
}

function ReleaseNotesLabel() {
  return (
    <Translate id="versionsPage.versionEntry.releaseNotes">
      Release Notes
    </Translate>
  );
}


// getVersionedDocs is getting all the GO Feature Flag Versions directly from Github release.
async function getVersionedDocs(owner, repo, path) {
  const apiUrl = `https://api.github.com/repos/${owner}/${repo}/contents/${path}`;
  const response = await fetch(apiUrl);

  if (!response.ok) {
    throw new Error(`GitHub API responded with status code: ${response.status}`);
  }

  const directoryListing = await response.json();
  const versionDirectories = directoryListing
    .filter(item => item.type === 'dir' && item.name.startsWith('version-'))
    .map(item => {
      const version = item.name.replace('version-', '')
      return {
        version,
        link: item.html_url,
      }
    })
    // Custom sort in descending order by version numbers
    .sort((a, b) => {
      const versionA = a.version.substring(1).split('.').map(Number); // Exclude 'v' and parse numbers
      const versionB = b.version.substring(1).split('.').map(Number);
      for (let i = 0; i < Math.max(versionA.length, versionB.length); i++) {
        if ((versionA[i] || 0) > (versionB[i] || 0)) return -1;
        if ((versionA[i] || 0) < (versionB[i] || 0)) return 1;
      }
      return 0;
    });

  return versionDirectories;
}


// This component is displaying all the versions of GO Feature Flag from GitHub release.
const VersionList = () => {
  const [versions, setVersions] = useState([]);
  const docusaurusVersions = useVersions(docsPluginId);

  useEffect(() => {
    const fetchVersions = async () => {
      try {
        const versionData = await getVersionedDocs('thomaspoignant', 'go-feature-flag', 'website/versioned_docs');
        setVersions(versionData);
      } catch (error) {
        console.error(error);
      }
    };

    fetchVersions();
  }, []);

  function isDocusaurusVersion(version) {
    return docusaurusVersions.some(docusaurusVersion => docusaurusVersion.name === version);
  }

  function getDocusaurusVersionLink(version) {
    return docusaurusVersions.find(docusaurusVersion => docusaurusVersion.name === version).path;
  }

  const {
    siteConfig: { organizationName, projectName }
  } = useDocusaurusContext();

  return (
    <div className="margin-bottom--lg">
      <Heading as="h3" id="archive">
        <Translate id="versionsPage.archived.title">
          All available Versions
        </Translate>
      </Heading>
      <p>
        <Translate id="versionsPage.archived.description">
          Here you can find documentation for previous versions of GO Feature Flag.
        </Translate>
      </p>
      <table>
        <tbody>
        {versions.map((version, index) => (
          <tr key={version.version}>
            <th>{version.version}</th>
              <td>
                {isDocusaurusVersion(version.version) ?
                  <Link to={getDocusaurusVersionLink(version.version)}><DocumentationLabel /></Link> :
                  <Link to={version.link}><DocumentationLabel />&nbsp;
                    <i className="fa fa-external-link" aria-hidden="true"></i>
                  </Link>}
              </td>
            <td>
            <Link href={`${getRepoURL(organizationName, projectName)}/releases/tag/${version.version}`}>
                  <ReleaseNotesLabel />
                </Link>
              </td>
            </tr>
        ))}
        </tbody>
      </table>
    </div>);
};


const getRepoURL= (organizationName, projectName) =>{
  return `https://github.com/${organizationName}/${projectName}`;
}

export default function Version() {
  const versions = useVersions(docsPluginId);
  const latestVersion = useLatestVersion(docsPluginId);
  const currentVersion = versions.find(
    (version) => version.name === "current"
  );

  const {
    siteConfig: { organizationName, projectName }
  } = useDocusaurusContext();

  return (
    <Layout
      title="Versions"
      description="GO Feature Flag Versions page listing all versions">
      <main className="container margin-vert--lg">
        <Heading as="h1">
          <Translate id="versionsPage.title">
            GO Feature Flag documentation versions
          </Translate>
        </Heading>

        <div className="margin-bottom--lg">
          <Heading as="h3" id="next">
            <Translate id="versionsPage.current.title">
              Current version (Stable)
            </Translate>
          </Heading>
          <p>
            <Translate id="versionsPage.current.description">
              Here you can find the documentation for current released version.
            </Translate>
          </p>
          <table>
            <tbody>
            <tr>
              <th>{latestVersion.label}</th>
              <td>
                <Link to={latestVersion.path}>
                  <DocumentationLabel />
                </Link>
              </td>
              <td>
                <Link to={`${getRepoURL(organizationName,projectName)}/releases/tag/${latestVersion.name}`}>
                  <ReleaseNotesLabel />
                </Link>
              </td>
            </tr>
            </tbody>
          </table>
        </div>


        {currentVersion !== latestVersion && (
          <div className="margin-bottom--lg">
            <Heading as="h3" id="latest">
              <Translate id="versionsPage.next.title">
                Next version (Unreleased)
              </Translate>
            </Heading>
            <p>
              <Translate id="versionsPage.next.description">
                Here you can find the documentation for work-in-process
                unreleased version.
              </Translate>
            </p>
            <table>
              <tbody>
              <tr>
                <th>{currentVersion.label}</th>
                <td>
                  <Link to={currentVersion.path}>
                    <DocumentationLabel />
                  </Link>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        )}


        <VersionList />
      </main>
    </Layout>
  );
}

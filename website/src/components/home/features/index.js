import React from 'react';
import styles from './styles.module.css';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import PropTypes from 'prop-types';
import {integrations} from '@site/data/integrations';
import {sdk} from '@site/data/sdk';

SocialIcon.propTypes = {
  colorClassName: PropTypes.string,
  fontAwesomeIcon: PropTypes.string,
  img: PropTypes.string,
  tooltipText: PropTypes.string.isRequired,
  backgroundColor: PropTypes.string,
};

function SocialIcon(props) {
  return (
    <div className={styles.tooltip}>
      <span
        className={clsx(styles.socialIcon, props.colorClassName)}
        style={{backgroundColor: props.backgroundColor}}>
        {props.fontAwesomeIcon && <i className={props.fontAwesomeIcon}></i>}
        {props.img && !props.fontAwesomeIcon && (
          <img src={props.img} width="36" alt={'logo'} />
        )}
      </span>
      <span className={styles.tooltiptext}>{props.tooltipText}</span>
    </div>
  );
}

export function Rollout() {
  return (
    <div className="container my-8">
      <div className="row">
        <div className={'col col--6'}>
          <div className={clsx(styles.featureContent, 'mr-25')}>
            <h2>Advanced rollout capabilities</h2>
            <p>
              Feature flags allows to unlink deploy and release, this means that
              you can decide how to release a feature without thinking about
              architecture and complex deployments.
            </p>
            <p>
              These capabilities will give you more control on your rollout
              changes and will ensure that everything happened as expected.
            </p>
            <div className={styles.featureContentList}>
              <ul>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>Rules: Impact only the users you want to target.</p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    Canary release: Enable the feature only to a subset of your
                    users.
                  </p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    Progressive rollout: Affect from 0% to 100% of users within
                    a time frame, you can monitor while releasing the change.
                  </p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    Scheduled Workflows: Modify your flag at a specific time to
                    impact more/less users.
                  </p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    A/B testing: Split your audience in multiple groups and
                    track their usage.
                  </p>
                </li>
              </ul>
            </div>
            <p className={'mt-3'}>
              <Link to={'/docs/configure_flag/rollout-strategies'}>
                {' '}
                See our rollout capabilites{' '}
                <i className="fa-solid fa-arrow-right"></i>
              </Link>
            </p>
          </div>
        </div>
        <div className={'col col--6'}>
          <div className={clsx(styles.imgRollout, 'max-md:hidden')}>
            <div className={styles.featureImage}>
              <img src="img/features/rollout.png" alt="feature-image" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export function Sdk() {
  return (
    <div className="container my-8">
      <div className="row">
        <div className={'col col--6'}>
          <div className={'grid grid-cols-4'}>
            {sdk.map(sdk => (
              <SocialIcon
                key={sdk.name}
                backgroundColor="#cdf7e7"
                fontAwesomeIcon={sdk.faLogo}
                tooltipText={sdk.name}
              />
            ))}
          </div>
        </div>
        <div className="col col--6">
          <div className={clsx(styles.featureContent, 'mr-25')}>
            <h2>Supports your favorite languages</h2>
            <p>
              GO Feature Flag believe in OpenSource, and offer providers for the
              feature flag standard{' '}
              <Link href={'https://openfeature.dev'}>OpenFeature</Link>.
              <br />
              In combination with the <b>Open Feature SDKs</b> these{' '}
              <b>providers</b> will allow you to use GO Feature Flag with all
              supported languages.
            </p>
            <p>
              <Link to={'/docs/sdk'}>
                {' '}
                See our SDKs <i className="fa-solid fa-arrow-right"></i>
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export function Integration() {
  const allIntegrations = [
    ...integrations.retrievers,
    ...integrations.exporters,
    ...integrations.notifiers,
  ];
  const displayIntegrations = allIntegrations
    .map(({name, logo, faLogo, bgColor}) => ({name, logo, faLogo, bgColor}))
    .filter(
      // remove duplicates
      (integration, index, self) =>
        index ===
        self.findIndex(
          t =>
            t.name === integration.name &&
            t.logo === integration.logo &&
            t.faLogo === integration.faLogo &&
            t.bgColor === integration.bgColor
        )
    );

  return (
    <div className="container my-8">
      <div className="row">
        <div className={'col col--6'}>
          <div className={clsx(styles.featureContent, 'mr-25')}>
            <h2>Integrates with different systems</h2>
            <p>
              GO Feature flag is cloud ready and can retrieve its configuration
              from various places, store your data usage where you want and
              notify you when something changes.
            </p>
            <div className={styles.featureContentList}>
              <ul>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    Retrieve your file from S3, Google Cloud, Github,
                    Kubernetes, and more.
                  </p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>Store flags usage in your favorite dataset (S3, GCP, ).</p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>
                    Be notified on slack or via a webhook that your flag has
                    changed.
                  </p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>You can also extend GO Feature Flag if needed.</p>
                </li>
              </ul>
              <p className={'mt-10 flex gap-2 items-center'}>
                {' '}
                See our integrations <i className="fa-solid fa-arrow-right"></i>
                <Link to={'/docs/integrations/store-flags-configuration'}>
                  Retrievers
                </Link>
                |
                <Link to={'/docs/integrations/export-evaluation-data'}>
                  Exporters
                </Link>
                |
                <Link to={'/docs/integrations/notify-flags-changes'}>
                  Notifiers
                </Link>
              </p>
            </div>
          </div>
        </div>
        <div className={'col col--6'}>
          <div className={'grid grid-cols-4'}>
            {displayIntegrations.map(integration => (
              <SocialIcon
                key={integration.name}
                backgroundColor={integration.bgColor}
                fontAwesomeIcon={integration.faLogo}
                img={integration.logo}
                tooltipText={integration.name}
                colorClassName={''}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export function OpenFeatureEcosystem() {
  // how to know if we are in dark mode
  return (
    <div className="container my-8">
      <div className="row">
        <div className={'col col--6'}>
          <div className={clsx(styles.openfeaturelogo, 'text-center xl:mt-16')}>
            <img
              src={'img/features/openfeature.svg'}
              alt="openfeature-logo"
              className={styles.openfeaturelogo}
            />
          </div>
        </div>
        <div className={'col col--6'}>
          <div className={clsx(styles.featureContent, 'mr-25')}>
            <h2>Part of the OpenFeature Ecosystem</h2>
            <p>
              At GO Feature Flag, we believe in the power of open standards and,
              the importance of vendor neutrality. That's why we've chosen to
              rely on Open Feature for our SDKs.
              <br />
              By adopting GO Feature Flag you embrace the OpenFeature standard
              and you get all the benefits of the ecosystem.
            </p>
            <div className={styles.featureContentList}>
              <ul>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>Open-Source SDKs</p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>No Vendor Lock-In</p>
                </li>
                <li>
                  <i className="fa-solid fa-circle-arrow-right"></i>
                  <p>OpenFeature community based support for SDKs</p>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

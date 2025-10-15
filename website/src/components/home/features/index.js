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

FeatureListItem.propTypes = {
  children: PropTypes.node.isRequired,
  icon: PropTypes.string,
};

function FeatureListItem({children, icon = 'fa-solid fa-circle-arrow-right'}) {
  return (
    <li className="flex text-left gap-2">
      <i className={`${icon} text-titles-500 text-2xl`}></i>
      <p className="dark:text-gray-50 font-poppins text-[1rem] font-[500] pt-1">
        {children}
      </p>
    </li>
  );
}

FeatureList.propTypes = {
  children: PropTypes.node.isRequired,
};

function FeatureList({children}) {
  return (
    <div>
      <ul className="list-none text-left pl-0">{children}</ul>
    </div>
  );
}

FeatureDescriptionBlock.propTypes = {
  childrens: PropTypes.node.isRequired,
};
function FeatureDescriptionBlock({childrens = []}) {
  return (
    <div>
      {childrens.map((children, idx) => (
        <p
          key={idx}
          className="text-[1.05rem] text-gray-500 font-poppins font-[400] leading-8">
          {children}
        </p>
      ))}
    </div>
  );
}

export function Rollout() {
  return (
    <div className="container my-8">
      <div className="row">
        <div className={'col col--6'}>
          <div className="mr-25">
            <h2 className="m-t-[2rem] text-4xl font-poppins font-[800] text-left tracking-[-0.08rem] color-gray-50">
              Advanced rollout capabilities
            </h2>
            <FeatureDescriptionBlock
              childrens={[
                'Feature flags allow you to unlink deploy and release. This means you can decide how to release a feature without worrying about architecture and complex deployments.',
                'These capabilities give you more control over your rollout changes and ensure that everything happens as expected.',
              ]}
            />
            <FeatureList>
              <FeatureListItem>
                Rules: Impact only the users you want to target.
              </FeatureListItem>
              <FeatureListItem>
                Canary release: Enable the feature only to a subset of your
                users.
              </FeatureListItem>
              <FeatureListItem>
                Progressive rollout: Affect from 0% to 100% of users within a
                time frame, you can monitor while releasing the change.
              </FeatureListItem>
              <FeatureListItem>
                Scheduled Workflows: Modify your flag at a specific time to
                impact more/less users.
              </FeatureListItem>
              <FeatureListItem>
                A/B testing: Split your audience in multiple groups and track
                their usage.
              </FeatureListItem>
            </FeatureList>
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
          <div className="mr-25">
            <h2 className="m-t-[2rem] text-4xl font-poppins font-[800] text-left tracking-[-0.08rem] color-gray-50">
              Supports your favorite languages
            </h2>
            <FeatureDescriptionBlock
              childrens={[
                <>
                  GO Feature Flag believe in OpenSource, and offer providers for
                  the feature flag standard{' '}
                  <Link href={'https://openfeature.dev'}>OpenFeature</Link>.
                  <br />
                  In combination with the <b>Open Feature SDKs</b> these{' '}
                  <b>providers</b> will allow you to use GO Feature Flag with
                  all supported languages.
                </>,
              ]}
            />
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
          <div className="mr-25">
            <h2 className="m-t-[2rem] text-4xl font-poppins font-[800] text-left tracking-[-0.08rem] color-gray-50">
              Integrates with different systems
            </h2>
            <FeatureDescriptionBlock
              childrens={[
                'GO Feature flag is cloud ready and can retrieve its configuration from various places, store your data usage where you want and notify you when something changes.',
              ]}
            />
            <FeatureList>
              <FeatureListItem>
                Retrieve your file from S3, Google Cloud, Github, Kubernetes,
                and more.
              </FeatureListItem>
              <FeatureListItem>
                Store flags usage in your favorite dataset (S3, GCP, and many
                more ...)
              </FeatureListItem>
              <FeatureListItem>
                Be notified on slack or via a webhook that your flag has
                changed.
              </FeatureListItem>
              <FeatureListItem>
                You can also extend GO Feature Flag if needed.
              </FeatureListItem>
            </FeatureList>
            <div>
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
          <div className="mr-25">
            <h2 className="m-t-[2rem] text-4xl font-poppins font-[800] text-left tracking-[-0.08rem] color-gray-50">
              Part of the OpenFeature Ecosystem
            </h2>
            <FeatureDescriptionBlock
              childrens={[
                "At GO Feature Flag, we believe in the power of open standards and, the importance of vendor neutrality. That's why we've chosen to rely on Open Feature for our SDKs.",
                'By adopting GO Feature Flag you embrace the OpenFeature standard and you get all the benefits of the ecosystem.',
              ]}
            />
            <FeatureList>
              <FeatureListItem>Open-Source SDKs</FeatureListItem>
              <FeatureListItem>No Vendor Lock-In</FeatureListItem>
              <FeatureListItem>
                OpenFeature community based support for SDKs
              </FeatureListItem>
            </FeatureList>
          </div>
        </div>
        <div className={'col col--6'}>
          <div className="flex text-center xl:mt-24 pt-8 md:pt-0 align-middle justify-center">
            <img
              src={'img/features/openfeature.svg'}
              alt="openfeature-logo"
              className="max-w-lg"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

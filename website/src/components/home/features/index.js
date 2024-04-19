import React from 'react';
import styles from './styles.module.css';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import PropTypes from 'prop-types';
import sqslogo from '@site/static/docs/collectors/sqs.png';
import pubsublogo from '@site/static/docs/collectors/pubsub.png'
import s3logo from '@site/static/docs/collectors/s3.png';
import webhooklogo from '@site/static/docs/collectors/webhook.png';
import {Headline} from '../headline';

SocialIcon.propTypes = {
  colorClassName: PropTypes.string.isRequired,
  fontAwesomeIcon: PropTypes.string,
  img: PropTypes.string,
  tooltipText: PropTypes.string.isRequired,
};
function SocialIcon(props) {
  return (
    <div className="col-1-4 mobile-col-1-2">
      <div className={styles.tooltip}>
        <span className={clsx(styles.socialIcon, props.colorClassName)}>
          {props.fontAwesomeIcon && <i className={props.fontAwesomeIcon}></i>}
          {props.img && <img src={props.img} width="36" />}
        </span>
        <span className={styles.tooltiptext}>{props.tooltipText}</span>
      </div>
    </div>
  );
}

function Rollout() {
  return (
    <div className="grid grid-pad">
      <div className="col-1-2">
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
                  Progressive rollout: Affect from 0% to 100% of users within a
                  time frame, you can monitor while releasing the change.
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
                  A/B testing: Split your audience in multiple groups and track
                  their usage.
                </p>
              </li>
            </ul>
          </div>
        </div>
      </div>
      <div className={clsx('col-1-2', styles.imgRollout)}>
        <div className={styles.featureImage}>
          <img
            src="img/features/rollout.png"
            alt="feature-image"
            id="shape-01"
          />
        </div>
      </div>
    </div>
  );
}

function OpenFeature() {
  return (
    <div className="grid grid-pad">
      <div className="col-1-2">
        <div className={clsx('grid grid-pad', styles.soc)}>
          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-go-original-wordmark colored"
            tooltipText="GO"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-java-plain-wordmark colored"
            tooltipText="Java"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-kotlin-plain-wordmark colored"
            tooltipText="Kotlin"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-javascript-plain colored"
            tooltipText="Javascript"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-typescript-plain colored"
            tooltipText="Typescript"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-dot-net-plain-wordmark colored"
            tooltipText=".NET"
          />

          <SocialIcon
            colorClassName={styles.socialIconCyan}
            fontAwesomeIcon="devicon-android-plain colored"
            tooltipText="Kotlin / Android"
          />
        </div>
      </div>
      <div className="col-1-2">
        <div className={clsx(styles.featureContent, 'mr-25')}>
          <h2>Supports your favorite languages</h2>
          <p>
            GO Feature Flag believe in OpenSource and offer providers for the
            new standard for feature flags{' '}
            <Link href={'https://openfeature.dev'}>OpenFeature</Link>.
            <br />
            In combination with the <b>Open Feature SDKs</b> these{' '}
            <b>providers</b> will allow you to use GO Feature Flag with all
            supported languages.
          </p>
          <p>
            <Link to={'/docs/openfeature_sdk/sdk'}>
              {' '}
              See our SDKs <i className="fa-solid fa-arrow-right"></i>
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

function Integration() {
  return (
    <div className="grid grid-pad">
      <div className="col-1-2">
        <div className={clsx(styles.featureContent, 'mr-25')}>
          <h2>Integrates with different systems</h2>
          <p>
            GO Feature flag is cloud ready and can retrieve its configuration
            from various places, store your data usage where you want and notify
            you when something changes.
          </p>
          <div className={styles.featureContentList}>
            <ul>
              <li>
                <i className="fa-solid fa-circle-arrow-right"></i>
                <p>
                  Retrieve your file from S3, Google Cloud, Github, Kubernetes,
                  and more.
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
          </div>
        </div>
      </div>
      <div className="col-1-2">
        <div className={clsx('grid grid-pad', styles.soc)}>
          <SocialIcon
            colorClassName={styles.socialIconBlue}
            fontAwesomeIcon="devicon-kubernetes-plain"
            tooltipText="Kubernetes"
          />

          <SocialIcon
            colorClassName={styles.socialIconBlack}
            fontAwesomeIcon="fab fa-github fa-stack-1x fa-inverse"
            tooltipText="GitHub"
          />

          <SocialIcon
            colorClassName={styles.socialIconAws}
            img={s3logo}
            tooltipText="AWS S3"
          />

          <SocialIcon
            colorClassName={styles.socialIconPurple}
            fontAwesomeIcon="fab fa-slack fa-stack-1x fa-inverse"
            tooltipText="Slack"
          />
        </div>
        <div className={clsx('grid grid-pad', styles.soc)}>

          <SocialIcon
            colorClassName={styles.socialIconBlack}
            fontAwesomeIcon="fas fa-file fa-stack-1x fa-inverse"
            tooltipText="Local file"
          />

          <SocialIcon
            colorClassName={styles.socialIconGreen}
            fontAwesomeIcon="fas fa-cloud-arrow-down fa-stack-1x fa-inverse"
            tooltipText="HTTP endpoint"
          />

          <SocialIcon
            colorClassName={styles.socialIconBlue}
            fontAwesomeIcon="devicon-googlecloud-plain"
            tooltipText="Google Cloud storage"
          />

          <SocialIcon
            colorClassName={styles.socialIconGreen}
            img={webhooklogo}
            tooltipText="Webhooks"
          />
        </div>
        <div className={clsx('grid grid-pad', styles.soc)}>
          <SocialIcon
            colorClassName={styles.socialIconMongodb}
            fontAwesomeIcon="devicon-mongodb-plain-wordmark colored"
            tooltipText="Mongodb"
          />
          <SocialIcon
            colorClassName={styles.socialIconGitlab}
            fontAwesomeIcon="devicon-gitlab-plain colored"
            tooltipText="Gitlab"
          />
          <SocialIcon
            colorClassName={styles.socialIconAws}
            img={sqslogo}
            tooltipText="AWS SQS"
          />
          <SocialIcon
              colorClassName={styles.socialIconPubSub}
              img={pubsublogo}
              tooltipText="Google PubSub"
          />
          <SocialIcon
            colorClassName={styles.socialIconKafka}
            fontAwesomeIcon="devicon-apachekafka-original colored"
            tooltipText="Apache Kafka"
          />
          <SocialIcon
            colorClassName={styles.socialIconBlack}
            fontAwesomeIcon="devicon-redis-plain-wordmark colored"
            tooltipText="Redis"
          />
        </div>
      </div>
    </div>
  );
}

export function Features() {
  return (
    <section className={styles.feature}>
      <div className="container">
        <OpenFeature />
        <Integration />
        <Headline />
        <Rollout />
      </div>
    </section>
  );
}

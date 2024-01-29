import React from 'react';
import styles from './styles.module.css';
import clsx from 'clsx';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';

import chapatiSystems from '@site/static/img/using-it/chapati.systems.png';
import lyft from '@site/static/img/using-it/lyft.png';

export function UsingIt() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <section className={styles.usingit}>
      <div className="grid grid-pad">
        <div className={clsx('col-1-1')}>
          <span className={clsx(styles.mainTitle)}>
            Who is using GO Feature Flag?
          </span>
          <div className={clsx('grid grid-pad', styles.logo)}>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content"></div>
            </div>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content"></div>
            </div>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content">
                <Link to={'https://github.com/lyft/atlantis'}>
                  <img src={lyft} alt={'Lyft'} />
                </Link>
              </div>
            </div>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content">
                <img src={chapatiSystems} alt={'chapati systems'} />
              </div>
            </div>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content"></div>
            </div>
            <div className={'col-1-6 mobile-col-1-1'}>
              <div className="content"></div>
            </div>
          </div>
        </div>
      </div>
      <div className={styles.contactUs}>
        Want to be listed here?{' '}
        <Link to={'mailto:contact@gofeatureflag.org'}>Contact us</Link>
      </div>
    </section>
  );
}

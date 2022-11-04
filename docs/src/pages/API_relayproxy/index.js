import { RedocStandalone } from 'redoc';
import React from 'react';
import Layout from '@theme/Layout';
import styles from './index.module.css'
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

export default function redoc() {
  const {siteConfig} = useDocusaurusContext();
  return(
    <Layout title="Relay proxy endpoints documentation">
      <div className={styles.redocContainer}>
        <RedocStandalone
          specUrl={siteConfig.customFields.swaggerURL}
          options={{
            hideHostname: true,
            disableSearch: true,
            nativeScrollbars: false,
            pathInMiddlePanel: true,
            jsonSampleExpandLevel: 5,
          }}
        />
      </div>
    </Layout>
  );
}

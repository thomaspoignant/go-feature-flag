import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {Whatis} from '../components/home/whatis';
import {HomeHeader} from '../components/home/HomeHeader';
import {Benefit} from '../components/home/benefit';
import {
  Integration,
  OpenFeatureEcosystem,
  Rollout,
  Sdk,
} from '../components/home/features';
import {UsingIt} from '../components/home/using-it';
import {Headline} from '../components/home/headline';

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.tagline}`}
      description={`${siteConfig.customFields.description}`}>
      <HomeHeader />
      <Whatis />
      <UsingIt />
      <div className={'my-10'}></div>
      <OpenFeatureEcosystem />
      <Sdk />
      <Integration />
      <Headline />
      <Rollout />
      <Benefit />
    </Layout>
  );
}

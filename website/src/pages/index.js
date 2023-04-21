import React from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {Whatis} from '../components/home/whatis';
import {HomeHeader} from '../components/home/HomeHeader';
import {Benefit} from '../components/home/benefit';
import {Features} from '../components/home/features';

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title="A Simple Open Source Feature Flag Solution"
      description={`${siteConfig.tagline}`}>
      <HomeHeader />
      <Whatis />
      <Features />
      <Benefit />
    </Layout>
  );
}

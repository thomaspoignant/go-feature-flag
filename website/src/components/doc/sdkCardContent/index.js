import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';
import PropTypes from 'prop-types';

SdkCardContent.prototype = {
  features: PropTypes.array.isRequired,
};
export function SdkCardContent(props) {
  return (
    <>
      {featureIcon(props.features, 'remoteEval')} Remote evaluation <br />
      {featureIcon(props.features, 'localCache')} Local cache
      <br />
      {featureIcon(props.features, 'dynamicRefresh')} Dynamic cache refresh
      <br />
    </>
  );
}

function featureIcon(features, key) {
  if (features.includes(key)) {
    return <i className={clsx('fa-solid fa-circle-check', styles.green)}></i>;
  }
  return <i className={clsx('fa-solid fa-person-digging', styles.orange)}></i>;
}

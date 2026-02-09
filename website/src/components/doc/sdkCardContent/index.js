import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';
import PropTypes from 'prop-types';

export function SdkCardContent(props) {
  return (
    <>
      {props.badgeUrl && (
        <>
          <img alt="badge" src={props.badgeUrl} />
          <br />
        </>
      )}
      {featureIcon(props.features, 'remoteEval')} Remote evaluation <br />
      {featureIcon(props.features, 'localCache')} Local cache
      <br />
      {featureIcon(props.features, 'dynamicRefresh')} Dynamic cache refresh
      <br />
    </>
  );
}

SdkCardContent.propTypes = {
  features: PropTypes.array.isRequired,
  badgeUrl: PropTypes.string.isRequired,
};

function featureIcon(features, key) {
  if (features.includes(key)) {
    return <i className={clsx('fa-solid fa-circle-check', styles.green)}></i>;
  }
  return <i className={clsx('fa-solid fa-person-digging', styles.orange)}></i>;
}

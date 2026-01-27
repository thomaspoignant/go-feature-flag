import React from 'react';
import PropTypes from 'prop-types';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';

export function ConfigCardContent(props) {
  return (
    <div className={styles.linkBox}>
      {props.relayproxyLink && (
        <Link to={props.relayproxyLink}>
          <button className={styles.button}>
            <i className="fa-solid fa-server"></i> Configure the Relay Proxy
          </button>
        </Link>
      )}
      {props.goModuleLink && (
        <Link to={props.goModuleLink}>
          <button className={styles.button}>
            <i className="devicon-go-original-wordmark"></i> Configure the GO
            Module
          </button>
        </Link>
      )}
    </div>
  );
}

ConfigCardContent.propTypes = {
  goModuleLink: PropTypes.string,
  relayproxyLink: PropTypes.string,
};

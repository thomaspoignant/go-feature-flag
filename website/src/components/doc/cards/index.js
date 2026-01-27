import React from 'react';
import clsx from 'clsx';
import PropTypes from 'prop-types';
import styles from '@site/src/components/doc/cards/styles.module.css';
import Link from '@docusaurus/Link';

export function Cards(props) {
  const listItems = props.test.map((item, index) => (
    <Card {...item} key={index} />
  ));

  return <div className="grid grid-pad">{listItems}</div>;
}

function Card(item) {
  return (
    <div className={clsx('col-1-3 mobile-col-1-2', styles.container)}>
      <div className={styles.card}>
        <img src={item.logo} className={styles.cardLogo} />
        <div className={styles.cardDetails}>
          <div className={styles.title}>{item.name}</div>
          <div className={styles.linkBox}>
            {item.relayproxy && (
              <Link to={item.relayproxy}>
                <button className={clsx(styles.button)}>
                  <i className="fa-solid fa-server"></i> Configure the Relay
                  Proxy
                </button>
              </Link>
            )}
            {item.gomodule && (
              <Link to={item.gomodule}>
                <button className={clsx(styles.button)}>
                  <i className="devicon-go-original-wordmark"></i> Configure the
                  GO Module
                </button>
              </Link>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

Cards.propTypes = {
  test: PropTypes.array.isRequired,
};

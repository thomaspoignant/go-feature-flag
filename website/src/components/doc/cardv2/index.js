import React from 'react';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';
import clsx from 'clsx';
import PropTypes from 'prop-types';

Cards.prototype = {
  cards: PropTypes.array.isRequired,
};
export function Cards(props) {
  const listItems = props.cards.map((item, index) => (
    <Card {...item} key={index} />
  ));
  return <div className="grid grid-pad">{listItems}</div>;
}

Card.propTypes = {
  title: PropTypes.string.isRequired,
  badges: PropTypes.array,
  warningBadges: PropTypes.array,
  logoCss: PropTypes.string,
  logoImg: PropTypes.string,
  docLink: PropTypes.string,
};
export function Card(props) {
  return (
    <div className={clsx('col-1-3 mobile-col-1-1')}>
      <Link to={props.docLink} className={styles.link}>
        <div className={styles.card}>
          <div className={styles.header}>
            <span className={styles.socialIcon}>
              {props.logoCss && <i className={props.logoCss}></i>}
              {props.logoImg && (
                <img src={props.logoImg} className={styles.logoImg} />
              )}
            </span>
          </div>
          <div>
            <p className={styles.name}>{props.title}</p>
          </div>
          <p className={styles.message}>{props.content}</p>
          <div className={styles.badgeSection}>
            {props.badges &&
              props.badges.map((item, i) => {
                return (
                  <span
                    className={clsx(styles.badge, styles.badgeInfo)}
                    key={item}>
                    {item}
                  </span>
                );
              })}
            {props.warningBadges &&
              props.warningBadges.map((item, i) => {
                return (
                  <span
                    className={clsx(styles.badge, styles.badgeWarning)}
                    key={item}>
                    {item}
                  </span>
                );
              })}
          </div>
        </div>
      </Link>
    </div>
  );
}

import React from 'react';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';
import clsx from 'clsx';
import PropTypes from 'prop-types';

export function Cards(props) {
  const listItems = props.cards.map((item, index) => (
    <Card {...item} key={index} />
  ));
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 xl:grid-cols-3">
      {listItems}
    </div>
  );
}

Cards.propTypes = {
  cards: PropTypes.array.isRequired,
};

Card.propTypes = {
  title: PropTypes.string.isRequired,
  badges: PropTypes.array,
  warningBadges: PropTypes.array,
  logoCss: PropTypes.string,
  logoImg: PropTypes.string,
  docLink: PropTypes.string,
  content: PropTypes.string,
};

export function Card(props) {
  return (
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
            props.badges.map(item => {
              return (
                <span
                  className={clsx(styles.badge, styles.badgeInfo)}
                  key={item}>
                  {item}
                </span>
              );
            })}
          {props.warningBadges &&
            props.warningBadges.map(item => {
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
  );
}

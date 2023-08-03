import React from 'react';
import Link from "@docusaurus/Link";
import styles from './styles.module.css';
import clsx from "clsx";
import PropTypes from "prop-types";

Cards.prototype = {
  cards: PropTypes.array.isRequired
}
export function Cards(props) {
  const listItems = props.cards.map((item, index) =>  <Card {...item} key={index} />);
  return(
    <div className="grid grid-pad">
      {listItems}
    </div>
  );
}

Card.propTypes = {
  language: PropTypes.string.isRequired,
  badges: PropTypes.array.isRequired,
  warningBadges: PropTypes.array,
  features: PropTypes.array.isRequired,
  logo: PropTypes.string.isRequired,
  docLink: PropTypes.string.isRequired,
};
export function Card(props) {
  return (
    <div className={clsx("col-1-3 mobile-col-1-2")}>
      <Link to={props.docLink} className={styles.link}>
        <div className={styles.card}>
          <div className={styles.header}>
            <span className={styles.socialIcon}>
              <i className={props.logo}></i>
            </span>
            <div>
              <p className={styles.name}>{props.language}</p>
            </div>
          </div>
          <p className={styles.message}>
            {featureIcon(props.features, "remoteEval")} Remote evaluation <br />
            {featureIcon(props.features, "localCache")} Local cache<br />
            {featureIcon(props.features, "dynamicRefresh")} Dynamic cache refresh<br />
          </p>
          <div className={styles.badgeSection}>
            {props.badges.map((item, i) => {
              return (<span className={clsx(styles.badge, styles.badgeInfo)} key={item}>{item}</span>)
            })}
            {props.warningBadges && props.warningBadges.map((item, i) => {
              return (<span className={clsx(styles.badge, styles.badgeWarning)} key={item}>{item}</span>)
            })}
          </div>
        </div>
      </Link>
    </div>
  );
}


function featureIcon(features, key) {
  if(features.includes(key)){
    return <i className={clsx("fa-solid fa-circle-check", styles.green)}></i>;
  }
  return <i className={clsx("fa-solid fa-person-digging", styles.orange)}></i>;
}
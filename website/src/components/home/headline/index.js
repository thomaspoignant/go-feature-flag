import React from 'react';
import styles from './styles.module.css';

export function Headline() {
  return (
    <div className={styles.headline}>
      <div className={styles.title}>
      <h1>Optimize Feature Launches with Advanced rollout capabilities.<br/>From <span className={styles.green}>Progressive Rollouts</span> to <span className={styles.purple}>Sequential Releases.</span></h1>
      </div>
    </div>
  );
}

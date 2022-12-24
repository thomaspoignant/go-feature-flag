import clsx from 'clsx';
import styles from './styles.module.css';
import React from 'react';
import {useFormContext} from 'react-hook-form';
import PropTypes from 'prop-types';

Switch.propTypes = {
  label: PropTypes.string.isRequired,
  displayText: PropTypes.string,
  defaultChecked: PropTypes.bool,
};
export function Switch({label, displayText, defaultChecked = false}) {
  const {register} = useFormContext();
  return (
    <div className={styles.container}>
      <div className={clsx(styles.displayText)}>{displayText}</div>
      <div>
        <div className={clsx(styles.switch, styles.r)}>
          <input
            type="checkbox"
            className={styles.checkbox}
            {...register(label)}
            defaultChecked={defaultChecked}
          />
          <div className={styles.knobs}></div>
          <div className={styles.layer}></div>
        </div>
      </div>
    </div>
  );
}

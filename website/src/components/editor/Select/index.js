import styles from './styles.module.css';
import React from 'react';
import {useFormContext} from 'react-hook-form';
import PropTypes from 'prop-types';

Select.propTypes = {
  title: PropTypes.string,
  content: PropTypes.array,
  label: PropTypes.string.isRequired,
  required: PropTypes.bool,
};
export function Select({title, content, label, required}) {
  const {register} = useFormContext();
  return (
    <div className={styles.selector}>
      <select defaultValue="0" {...register(label, required)}>
        <option disabled={true} defaultChecked={true} value={null}>
          {title}
        </option>
        {content.map((item, index) => (
          <option
            value={item.value}
            key={`variation_type_${item.value}_${item.displayText}`}>
            {item.displayName}
          </option>
        ))}
      </select>
    </div>
  );
}

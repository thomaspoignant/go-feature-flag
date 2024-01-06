import styles from './styles.module.css';
import React from 'react';
import {useFormContext} from 'react-hook-form';
import PropTypes from 'prop-types';

Select.propTypes = {
  title: PropTypes.string,
  content: PropTypes.array,
  label: PropTypes.string.isRequired,
  required: PropTypes.bool,
  controlled: PropTypes.bool,
};
export function Select({
  title,
  content,
  label,
  required,
  controlled = false,
  ...props
}) {
  const {register} = useFormContext();

  const registerProps = register(label, required);

  return (
    <div className={styles.selector}>
      <select
        defaultValue="0"
        {...(controlled ? {} : registerProps)}
        {...props}>
        <option disabled={true} defaultChecked={true} value={null}>
          {title}
        </option>
        {content.map(item => (
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

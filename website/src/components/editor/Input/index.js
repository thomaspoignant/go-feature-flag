import clsx from 'clsx';
import React from 'react';
import styles from './styles.module.css';
import {useFormContext} from 'react-hook-form';
import {ErrorMessage} from '@hookform/error-message';
import PropTypes from 'prop-types';

Input.propTypes = {
  label: PropTypes.string.isRequired,
  displayText: PropTypes.string,
  className: PropTypes.string,
  required: PropTypes.bool,
  type: PropTypes.string,
  validation: PropTypes.object,
  defaultValue: PropTypes.any,
  disablePlaceholder: PropTypes.bool,
  disableInlineErr: PropTypes.bool,
  controlled: PropTypes.bool,
};
export function Input({
  label,
  displayText,
  className,
  required,
  type,
  validation,
  defaultValue,
  controlled = false,
  disablePlaceholder = false,
  disableInlineErr = false,
  ...props
}) {
  const {register} = useFormContext();

  // inputType will return the input type based on the type provided in the props.
  function inputType() {
    switch (type) {
      case 'number':
        return 'number';
      default:
        return 'text';
    }
  }

  const registerProps = register(label, {
    required: {value: required, message: 'This field is required'},
    ...validation,
  });

  return (
    <div className={clsx(className || styles.editorInputContainer)}>
      <input
        id={`${label}.input`}
        defaultValue={defaultValue}
        className={styles.editorInput}
        type={inputType()}
        placeholder=" "
        {...(controlled ? {} : registerProps)}
        {...props}
      />
      {disablePlaceholder && <span>{displayText}</span>}
      <div className={styles.editorCut}></div>
      <label
        htmlFor={`${label}.input`}
        className={clsx(styles.editorPlaceholder)}>
        {displayText}
      </label>
      {!disableInlineErr && (
        <ErrorMessage name={label} render={inputErrorMessage} />
      )}
    </div>
  );
}

function inputErrorMessage({message}) {
  return <div className={styles.errorMessage}>{message}</div>;
}

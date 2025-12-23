import clsx from 'clsx';
import React from 'react';
import styles from './styles.module.css';
import {useFormContext} from 'react-hook-form';
import {ErrorMessage} from '@hookform/error-message';
import PropTypes from 'prop-types';

TextArea.propTypes = {
  label: PropTypes.string.isRequired,
  displayText: PropTypes.string,
  className: PropTypes.string,
  required: PropTypes.bool,
  validation: PropTypes.object,
  defaultValue: PropTypes.any,
  disablePlaceholder: PropTypes.bool,
  disableInlineErr: PropTypes.bool,
};
export function TextArea({
  label,
  displayText,
  className,
  required,
  validation,
  defaultValue,
  disablePlaceholder = false,
  disableInlineErr = false,
}) {
  const {register} = useFormContext();

  return (
    <div
      className={clsx(className || styles.editorTextAreaContainer)}>
      <textarea
        id={`${label}.textarea`}
        defaultValue={defaultValue}
        className={styles.editorTextArea}
        placeholder=" "
        {...register(label, {
          required: {value: required, message: 'This field is required'},
          ...validation,
        })}
      />
      {disablePlaceholder && <span>{displayText}</span>}
      <div className={styles.editorCut}></div>
      <label
        htmlFor={`${label}.textarea`}
        className={clsx(styles.editorPlaceholder)}>
        {displayText}
      </label>
      {!disableInlineErr && (
        <ErrorMessage name={label} render={textAreaErrorMessage} />
      )}
    </div>
  );
}

function textAreaErrorMessage({message}) {
  return <div className={styles.errorMessage}>{message}</div>;
}

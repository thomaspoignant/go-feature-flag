import React from 'react';
import {useFormContext} from 'react-hook-form';
import Highlight from 'react-highlight';
import 'highlight.js/styles/a11y-dark.css';
import YAML from 'yaml';
import * as TOML from '@iarna/toml';
import styles from './styles.module.css';
import {Select} from '../Select';
import { singleFlagFormConvertor } from "../utils";

function formToGoFeatureFlag(formData) {
  const goffFlags = {};
  formData.GOFeatureFlagEditor.filter(
    i => i.flagName !== undefined && i.flagName !== ''
  ).forEach(i => (goffFlags[i.flagName] = singleFlagFormConvertor(i)));

  return goffFlags;
}

function ErrorInForm() {
  return (
    <div className={styles.invalidForm}>
      Error in your configuration, please review the form.
    </div>
  );
}

export function FlagDisplay() {
  const {
    watch,
    formState: {errors},
  } = useFormContext();
  const data = watch();
  const isValid =
    errors &&
    Object.keys(errors).length === 0 &&
    Object.getPrototypeOf(errors) === Object.prototype;

  function formatFlagFile(config, format) {
    switch (format) {
      case 'json':
        return JSON.stringify(config, null, 2);
      case 'toml':
        return TOML.stringify(config);
      default:
        return YAML.stringify(config, null, 2);
    }
  }

  const select = [
    {value: 'yaml', displayName: 'YAML'},
    {value: 'json', displayName: 'JSON'},
    {value: 'toml', displayName: 'TOML'},
  ];

  return (
    <div className="col-4-12">
      {!isValid && <ErrorInForm />}
      {isValid && <Select
        title={'Format'}
        content={select}
        required={false}
        label={'flagFormat'}
      />}
      <div className={styles.space}></div>
      {isValid && (
        <Highlight className="JSON">
          {formatFlagFile(formToGoFeatureFlag(data), data.flagFormat)}
        </Highlight>
      )}
    </div>
  );
}

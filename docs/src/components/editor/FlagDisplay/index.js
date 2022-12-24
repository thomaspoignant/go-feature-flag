import React from 'react';
import {useFormContext} from 'react-hook-form';
import Highlight from 'react-highlight';
import 'highlight.js/styles/a11y-dark.css';
import YAML from 'yaml';
import * as TOML from '@iarna/toml';
import styles from './styles.module.css';
import {Select} from '../Select';

function formToGoFeatureFlag(formData) {
  function convertValueIntoType(value, type) {
    switch (type) {
      case 'json':
        try {
          return JSON.parse(value.value);
        } catch (e) {
          return undefined;
        }
      case 'number':
        return Number(value) || undefined;
      case 'boolean':
        if (typeof value == 'boolean') return value;
        return (
          value !== undefined &&
          (typeof value === 'string' || value instanceof String) &&
          value.toLowerCase() === 'true'
        );
      default:
        return String(value) || undefined;
    }
  }

  function convertRule(ruleForm) {
    let variation,
      percentage,
      progressiveRollout = undefined;
    const {selectedVar} = ruleForm;
    switch (selectedVar) {
      case 'percentage':
        percentage = {};
        ruleForm.percentages.forEach(i => (percentage[i.name] = i.value));
        break;
      case 'progressive':
        progressiveRollout = {
          initial: {
            variation: ruleForm.progressive.initial.selectedVar,
            percentage: ruleForm.progressive.initial.percentage || 0,
            date: ruleForm.progressive.initial.date,
          },
          end: {
            variation: ruleForm.progressive.end.selectedVar,
            percentage: ruleForm.progressive.end.percentage || 100,
            date: ruleForm.progressive.end.date,
          },
        };
        break;
      default:
        variation = selectedVar;
        break;
    }

    return {
      name: ruleForm.name || undefined,
      query: ruleForm.query,
      variation,
      percentage,
      progressiveRollout,
    };
  }
  function singleFlagFormConvertor(flagFormData) {
    const variationType = flagFormData.type;
    const variations = {};

    flagFormData.variations
      .filter(
        i =>
          i.name !== undefined &&
          i.name !== '' &&
          i.value !== undefined &&
          i.value !== ''
      )
      .forEach(
        i => (variations[i.name] = convertValueIntoType(i.value, variationType))
      );

    const targeting = flagFormData.targeting.map(t => convertRule(t));
    const trackEvents = convertValueIntoType(
      flagFormData.trackEvents,
      'boolean'
    );
    const disable = convertValueIntoType(flagFormData.disable, 'boolean');
    const defaultRule = convertRule(flagFormData.defaultRule);

    return {
      variations,
      disable: !disable ? undefined : disable,
      trackEvents: trackEvents ? undefined : trackEvents,
      version: flagFormData.version === '' ? undefined : flagFormData.version,
      targeting: targeting.length > 0 ? targeting : undefined,
      defaultRule,
    };
  }

  const goffFlags = {};
  formData.GOFeatureFlagEditor.filter(
    i => i.flagName !== undefined && i.flagName !== ''
  ).forEach(i => (goffFlags[i.flagName] = singleFlagFormConvertor(i)));

  return goffFlags;
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

  function ErrorInForm() {
    // TODO: best looking error here
    return (
      <div className={styles.invalidForm}>
        Error in your configuration, please review the form.
      </div>
    );
  }
  const select = [
    {value: 'yaml', displayName: 'YAML'},
    {value: 'json', displayName: 'JSON'},
    {value: 'toml', displayName: 'TOML'},
  ];

  return (
    <div className="col-4-12">
      {!isValid && <ErrorInForm />}
      <Select
        title={'Format'}
        content={select}
        required={false}
        label={'flagFormat'}
      />

      {isValid && (
        <Highlight className="JSON">
          {formatFlagFile(formToGoFeatureFlag(data), data.flagFormat)}
        </Highlight>
      )}
    </div>
  );
}

import React, {useState} from 'react';
import PropTypes from 'prop-types';
import {useFormContext} from 'react-hook-form';
import {singleFlagFormConvertor} from '../utils';
import {JsonEditor} from '../JsonEditor';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import clsx from 'clsx';
import styles from './styles.module.css';

export function FlagTest({flagInfo}) {
  const {siteConfig} = useDocusaurusContext();
  const {watch} = useFormContext();
  let result = watch();
  let keys = flagInfo.split('.');
  for (let key of keys) {
    result = result[key];
  }

  const flagName = result['flagName'];
  const flag = singleFlagFormConvertor(result);
  const defaultContext = {
    key: 'aae1cb41-c3cb-4753-a117-031ddc958e81',
    custom: {
      anonymous: true,
      firstname: 'John',
      lastname: 'Doe',
      email: 'john.doe@gofeatureflag.org',
      company: 'GO Feature Flag',
    },
  };

  const [data, setData] = useState({
    resolutionDetail: undefined,
    err: undefined,
  });

  const submit = async event => {
    event.preventDefault();
    try {
      const res = await fetch(siteConfig.customFields.playgroundEvaluationApi, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify({
          flagName: flagName,
          flag: flag,
          context: result.context.value ? JSON.parse(result.context.value) : {},
        }),
      });

      if (res.status === 200) {
        res.json().then(value => setData({resolutionDetail: value}));
      }
    } catch (e) {
      setData({err: e.toString()});
    }
  };

  return (
    <div>
      <h2>Test your flag</h2>
      <div className="grid grid-cols-12">
        <div className="col-span-5">
          <h4>Evaluation Context</h4>
          <JsonEditor
            label={`${flagInfo}.context`}
            value={JSON.stringify(defaultContext, ' ', 2)}
          />
        </div>
        <div className={clsx('col-span-2', styles.buttonContainer)}>
          <button
            className="pushy__btn pushy__btn--md pushy__btn--black"
            onClick={submit}>
            Evaluate Flag
          </button>
        </div>
        <div className="col-span-5">
          <h4>Flag evaluation Details</h4>
          <div>
            {data.resolutionDetail !== undefined && (
              <ul>
                <li>Value: {data.resolutionDetail.value.toString()}</li>
                <li>Variation Type: {data.resolutionDetail.variationType}</li>
                <li>Reason: {data.resolutionDetail.reason}</li>
                <li>Failed: {data.resolutionDetail.failed}</li>
                <li>Error Code: {data.resolutionDetail.errorCode}</li>
                <li>Track Events: {data.resolutionDetail.trackEvents}</li>
              </ul>
            )}
            {data.err !== undefined && (
              <span>Impossible to call the API: {data.err}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

FlagTest.propTypes = {
  flagInfo: PropTypes.string.isRequired,
};

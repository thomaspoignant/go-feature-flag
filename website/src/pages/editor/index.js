import React from 'react';
import {useForm, FormProvider, useFieldArray} from 'react-hook-form';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {FlagForm} from '../../components/editor/FlagForm';
import {FlagDisplay} from '../../components/editor/FlagDisplay';
import styles from '../../components/editor/Targeting/styles.module.css';
import clsx from 'clsx';

function App() {
  const EDITOR_NAME = 'GOFeatureFlagEditor';
  const methods = useForm({
    mode: 'onChange',
    defaultValues: {
      GOFeatureFlagEditor: [
        {
          flagName: 'my-first-flag',
          variations: [
            {name: 'Variation_1', value: true},
            {name: 'Variation_2', value: false},
          ],
          targeting: [],
          defaultRule: {},
        },
      ],
    },
  });

  const {fields, append} = useFieldArray({
    control: methods.control,
    name: EDITOR_NAME,
    rules: {minLength: 1},
  });

  const addNewFlag = event => {
    event.preventDefault();
    append({
      flagName: `new-flag-${fields.length}`,
      variations: [
        {name: 'Variation_1', value: true},
        {name: 'Variation_2', value: false},
      ],
      targeting: [],
      defaultRule: {},
    });
  };

  const onSubmit = event => {
    event.preventDefault();
  };

  return (
    <div className="grid-pad grid">
      <FormProvider {...methods}>
        <div className="col-8-12">
          <form onSubmit={methods.handleSubmit(onSubmit)}>
            {fields.map((field, index) => (
              <FlagForm label={`${EDITOR_NAME}.${index}`} key={field.id} />
            ))}
            <button className={styles.button} onClick={addNewFlag}>
              <span className="fa-stack fa-1x">
                <i
                  className={clsx(
                    'fa-solid fa-circle fa-stack-2x',
                    styles.bg
                  )}></i>
                <i className="fa-solid fa-plus fa-stack-1x fa-inverse"></i>
              </span>
              Add another flag
            </button>
          </form>
        </div>
        <FlagDisplay />
      </FormProvider>
    </div>
  );
}

export default function Page() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <App />
    </Layout>
  );
}

import React, {useEffect, useState} from 'react';
import {FormProvider, useFieldArray, useForm} from 'react-hook-form';
import Layout from '@theme/Layout';
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
    <div className="grid grid-cols-12 gap-4 my-4 ">
      <FormProvider {...methods}>
        <div className="col-span-8">
          <form onSubmit={methods.handleSubmit(onSubmit)}>
            {fields.map((field, index) => (
              <FlagForm label={`${EDITOR_NAME}.${index}`} key={field.id} />
            ))}
            <button
              type="button"
              className={styles.button}
              onClick={addNewFlag}>
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
        <div className={'col-span-4 ml-1'}>
          <FlagDisplay />
        </div>
      </FormProvider>
    </div>
  );
}

const STUDIO_CALLOUT_KEY = 'goff-studio-callout-dismissed';

function StudioCallout() {
  // Hidden until mount so the server render matches the first client render.
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    try {
      setVisible(localStorage.getItem(STUDIO_CALLOUT_KEY) !== 'true');
    } catch {
      setVisible(true);
    }
  }, []);

  const close = () => {
    setVisible(false);
    try {
      localStorage.setItem(STUDIO_CALLOUT_KEY, 'true');
    } catch {
      // storage unavailable, the callout comes back on next visit
    }
  };

  if (!visible) {
    return null;
  }

  return (
    <div className="mx-4 mt-4 flex items-start gap-3 rounded-lg border border-solid border-cyan-300 bg-cyan-50 p-4 text-gray-800 dark:border-cyan-800 dark:bg-cyan-950 dark:text-gray-100">
      <i className="fa-solid fa-flask mt-1 text-cyan-600 dark:text-cyan-400"></i>
      <p className="m-0 flex-1">
        <strong>Coming soon: GO Feature Flag Studio.</strong> This editor helps
        you write one flag file. Studio is the next step: a full-featured flag
        management platform for GO Feature Flag, where you create, edit and
        manage all your flags from a UI instead of editing files by hand.
        It&apos;s a work in progress, built in the open.{' '}
        <a
          href="https://github.com/go-feature-flag/studio"
          target="_blank"
          rel="noopener noreferrer"
          className="font-semibold text-cyan-800 dark:text-cyan-300">
          Follow the project on GitHub →
        </a>
      </p>
      <button
        type="button"
        aria-label="Close"
        onClick={close}
        className="cursor-pointer border-0 bg-transparent p-1 text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100">
        <i className="fa-solid fa-xmark"></i>
      </button>
    </div>
  );
}

export default function Page() {
  return (
    <Layout
      title="Edit Your Flag"
      description="GO Feature Flag editor is the simplest way to configure your feature flags.">
      <StudioCallout />
      <App />
    </Layout>
  );
}

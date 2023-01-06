import React from 'react';
import CodeEditor from '@uiw/react-textarea-code-editor';
import {useColorMode} from '@docusaurus/theme-common';
import styles from './styles.module.css';
import {useFormContext} from 'react-hook-form';
import PropTypes from 'prop-types';

JsonEditor.propTypes = {
  label: PropTypes.string.isRequired,
};
export function JsonEditor({label}) {
  const {register} = useFormContext();
  const {colorMode} = useColorMode();
  const {onChange, onBlur, name, ref} = register(`${label}.value`);
  return (
    <div className={styles.container} data-color-mode={colorMode}>
      <CodeEditor
        value=""
        language="json"
        placeholder=" Please enter JSON."
        padding={7}
        className={styles.container}
        onChange={onChange}
        onBlur={onBlur}
        name={name}
        ref={ref}
        style={{
          fontFamily:
            'ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace',
        }}
      />
    </div>
  );
}

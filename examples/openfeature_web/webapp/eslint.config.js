const { FlatCompat } = require('@eslint/eslintrc');
const js = require('@eslint/js');

// ESLint v9 uses "flat config" by default. This file keeps the previous
// `.eslintrc` behavior (env/extends/parser/rules) so webpack's ESLint plugin
// can find a config when running `npm run build` / `npm run watch`.
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

module.exports = [
  ...compat.config({
    env: {
      browser: true,
      es6: true,
    },
    extends: ['eslint:recommended'],
    globals: {
      Atomics: 'readonly',
      SharedArrayBuffer: 'readonly',
    },
    parser: '@babel/eslint-parser',
    parserOptions: {
      requireConfigFile: false,
      ecmaVersion: 2018,
      sourceType: 'module',
      allowImportExportEverywhere: true,
    },
    rules: {
      semi: ['error', 'always'],
      quotes: ['error', 'single'],
    },
  }),
];


import { defineConfig } from 'vite';
import { resolve } from 'node:path';

// PROVIDER_SRC=/path/to/js-sdk-contrib -> load the web provider from that checkout's source (HMR on provider edits)
// instead of the published npm package. Useful to test a provider branch that is not released yet.
const providerSrc = process.env.PROVIDER_SRC;

export default defineConfig({
  define: { __PROVIDER_SOURCE__: JSON.stringify(providerSrc ? `source: ${providerSrc}` : 'npm package') },
  resolve: providerSrc
    ? {
        alias: {
          '@openfeature/go-feature-flag-web-provider': resolve(providerSrc, 'libs/providers/go-feature-flag-web/src/index.ts'),
        },
        // The provider source has no node_modules: resolve its deps from this app (single OpenFeature singleton).
        dedupe: ['@openfeature/web-sdk', '@openfeature/core', 'copy-anything'],
      }
    : {},
  server: providerSrc ? { fs: { allow: ['.', providerSrc] } } : {},
});

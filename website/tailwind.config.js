const {fontFamily} = require('tailwindcss/defaultTheme');

/** @type {import('tailwindcss').Config} */
module.exports = {
  corePlugins: {
    preflight: false,
  },
  darkMode: ['class', '[data-theme="dark"]'],
  content: {
    files: ['./src/**/*.{js,jsx,ts,tsx,md,mdx}', './docs/**/*.{md,mdx}'],
    extract: {
      css: () => [], // Don't purge anything from CSS files
    },
  },
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Inter"', ...fontFamily.sans],
        jakarta: ['"Plus Jakarta Sans"', ...fontFamily.sans],
        mono: ['"Fira Code"', ...fontFamily.mono],
      },
      borderRadius: {
        sm: '4px',
      },
      colors: {
        goff: {
          50: '#edfcf7',
          100: '#cdf7e7',
          200: '#abefd9',
          300: '#74e1c4',
          400: '#3ccbaa',
          500: '#18b192',
          600: '#0c8f77',
          700: '#0a7263',
          800: '#0a5b4f',
          900: '#0a4a41',
          950: '#042a26',
        },
        titles: {
          500: '#9fbeb3',
        },
      },
    },
  },
  plugins: [],
};

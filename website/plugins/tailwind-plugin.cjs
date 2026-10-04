// Docusaurus minifies CSS with cssnano followed by clean-css. clean-css merges
// Tailwind v4's `@supports (color: color-mix(...))` blocks and moves them after
// the `dark:` variants, so e.g. `bg-gray-50/80` beats `dark:bg-[#161616]` in
// dark mode. This historical Docusaurus opt-out keeps cssnano only.
process.env.USE_SIMPLE_CSS_MINIFIER ??= "true";

function tailwindPlugin(context, options) {
  return {
    name: "tailwind-plugin",
    configurePostCss(postcssOptions) {
      postcssOptions.plugins = [
        require("@tailwindcss/postcss"),
        require("autoprefixer"),
      ];
      return postcssOptions;
    },
  };
}

module.exports = tailwindPlugin;

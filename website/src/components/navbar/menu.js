// Shared helpers to build multi-column navbar mega-menus (same approach as
// sdks.js: return a Tailwind-styled HTML string wired into the navbar as a
// `type: 'html'` dropdown item).
import React from 'react';
import {renderToStaticMarkup} from 'react-dom/server';

// Render a react-icons component to a static SVG string so it can be embedded
// in the HTML the dropdown is built from. Icons inherit the link color via
// `currentColor`. Returns an empty string when no icon is provided.
function renderIcon(icon, size = 16) {
  if (!icon) return '';
  return renderToStaticMarkup(
    React.createElement(icon, {size, 'aria-hidden': true})
  );
}

// Render a single menu link with a leading react-icons icon. `external` opens
// in a new tab with a trailing indicator.
function generateMenuLink({label, href, icon, external = false}) {
  const externalAttrs = external
    ? ' target="_blank" rel="noopener noreferrer"'
    : '';
  const externalIcon = external
    ? ' <i class="fa fa-external-link text-[0.65rem] align-baseline" aria-hidden="true"></i>'
    : '';
  return `
    <a
      href="${href}"${externalAttrs}
      class="no-underline hover:no-underline flex items-center gap-2.5 py-2 text-sm font-poppins font-[500] text-gray-800 dark:text-gray-200 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
    >
      <span class="inline-flex shrink-0">${renderIcon(icon)}</span>
      <span>${label}${externalIcon}</span>
    </a>
  `;
}

// Render a labelled column: a small indigo heading followed by its links.
// `withDivider` adds a subtle left border (used for every column after the first).
export function generateMenuColumn({title, items, withDivider = false}) {
  return `
    <div class="flex flex-col min-w-[180px] px-6 ${
      withDivider
        ? 'border-l border-t-0 border-r-0 border-b-0 border-solid border-gray-200 dark:border-gray-600'
        : ''
    }">
      <h3 class="text-xs font-poppins font-[600] uppercase tracking-wide text-indigo-500 dark:text-indigo-400 mb-2 mt-0">
        ${title}
      </h3>
      ${items.map(item => generateMenuLink(item)).join('')}
    </div>
  `;
}

// Render the GrowthBook-style "Explore" featured card: a gradient placeholder
// thumbnail, a title and an arrow, all wrapped in a single clickable link.
export function generateFeaturedCard({title, href, columnTitle = 'Explore'}) {
  return `
    <div class="flex flex-col min-w-[260px] px-6 border-l border-t-0 border-r-0 border-b-0 border-solid border-gray-200 dark:border-gray-600">
      <h3 class="text-xs font-poppins font-[600] uppercase tracking-wide text-indigo-500 dark:text-indigo-400 mb-2 mt-0">
        ${columnTitle}
      </h3>
      <a
        href="${href}"
        class="no-underline hover:no-underline flex items-center gap-4 rounded-xl bg-gray-100 dark:bg-gray-800 p-3 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
      >
        <span class="shrink-0 w-28 h-20 rounded-lg bg-gradient-to-br from-indigo-500 via-purple-500 to-teal-300"></span>
        <span class="flex flex-col gap-3">
          <span class="text-sm font-poppins font-[500] text-gray-800 dark:text-gray-200 leading-snug">
            ${title}
          </span>
          <i class="fa-solid fa-arrow-right text-gray-700 dark:text-gray-300"></i>
        </span>
      </a>
    </div>
  `;
}

// Render one "Get Help" card: a square icon tile on the left, a bold title and
// a muted description on the right. `iconSvg` is a raw inline SVG string.
export function generateHelpCard({
  title,
  description,
  href,
  icon,
  iconSvg,
  external = false,
}) {
  const externalAttrs = external
    ? ' target="_blank" rel="noopener noreferrer"'
    : '';
  const iconHtml = icon ? renderIcon(icon, 28) : iconSvg;
  return `
    <a
      href="${href}"${externalAttrs}
      class="no-underline hover:no-underline flex w-72 items-center gap-4 rounded-xl bg-gray-100 dark:bg-gray-800 p-3 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
    >
      <span class="shrink-0 w-14 h-14 rounded-lg bg-white dark:bg-gray-900 flex items-center justify-center text-gray-700 dark:text-gray-200">
        ${iconHtml}
      </span>
      <span class="flex flex-col gap-1 min-w-0">
        <span class="text-sm font-poppins font-[600] text-gray-900 dark:text-gray-100">
          ${title}
        </span>
        <span class="text-xs font-poppins font-[400] text-gray-500 dark:text-gray-400 leading-snug">
          ${description}
        </span>
      </span>
    </a>
  `;
}

// Render a labelled "Get Help" footer section: a heading above a row of cards,
// separated from the columns above by a horizontal divider.
export function generateHelpSection({title = 'Get Help', cards}) {
  return `
    <div class="mt-2 px-6 pt-3 border-t border-x-0 border-b-0 border-solid border-gray-200 dark:border-gray-600">
      <h3 class="text-xs font-poppins font-[600] uppercase tracking-wide text-indigo-500 dark:text-indigo-400 mb-2 mt-0">
        ${title}
      </h3>
      <div class="flex flex-row gap-3">
        ${cards.map(card => generateHelpCard(card)).join('')}
      </div>
    </div>
  `;
}

// Wrap a set of pre-rendered columns in the mega-menu container, mirroring the
// `.sdks-dropdown` styling used by sdks.js. An optional `footerHtml` (e.g. a
// "Get Help" section) is stacked below the columns row.
export function generateMegaMenu(columnsHtml, footerHtml = '') {
  if (!footerHtml) {
    return `
    <div class="megamenu-dropdown flex w-max flex-row flex-nowrap rounded-2xl py-2">
      ${columnsHtml}
    </div>
  `;
  }
  return `
    <div class="megamenu-dropdown flex w-max flex-col rounded-2xl py-2">
      <div class="flex flex-row flex-nowrap">
        ${columnsHtml}
      </div>
      ${footerHtml}
    </div>
  `;
}

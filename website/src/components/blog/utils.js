// Shared helpers for the blog card grid (`@theme/BlogListPage`).

// Closed list of categories, in the order the filter chips render them.
// Keep in sync with website/scripts/check-blog-frontmatter.mjs, which fails the
// build when a post declares a `category` that is not in this list.
export const CATEGORIES = ['Product', 'Engineering', 'Community', 'Case study'];

// Used when a post has no cover of its own. The CI check makes this rare, but a
// grid must never render an empty cell.
export const DEFAULT_COVER = '/img/logo/x-card.png';
export const DEFAULT_COVER_ALT = 'GO Feature Flag';

// Categories are matched against a URL query param, so 'Case study' travels as
// 'case-study'.
export const slugify = value =>
  value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');

/**
 * Flattens one `BlogListPage` item into the shape the cards consume.
 *
 * `assets.image` is the webpack-processed form of a relative front matter
 * `image:` (e.g. `./banner.png`); `frontMatter.image` covers absolute values
 * such as `/img/logo/x-card.png`. This is the same resolution order Docusaurus
 * itself uses for `og:image` and for the blog structured data, so one front
 * matter field drives the card, the hero and the social card.
 */
export const normalizePost = item => {
  const {metadata, frontMatter, assets} = item.content;
  return {
    permalink: metadata.permalink,
    title: metadata.title,
    description: metadata.description,
    date: metadata.date,
    formattedDate: metadata.formattedDate,
    readingTime: metadata.readingTime,
    authors: metadata.authors ?? [],
    image: assets?.image ?? frontMatter.image ?? DEFAULT_COVER,
    imageAlt: frontMatter.image_alt ?? DEFAULT_COVER_ALT,
    category: CATEGORIES.includes(frontMatter.category)
      ? frontMatter.category
      : null,
  };
};

export const formatReadingTime = readingTime =>
  typeof readingTime === 'number' ? `${Math.ceil(readingTime)} min read` : null;

import React, {useEffect, useMemo, useState} from 'react';
import PropTypes from 'prop-types';
import clsx from 'clsx';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {
  PageMetadata,
  HtmlClassNameProvider,
  ThemeClassNames,
} from '@docusaurus/theme-common';
import Layout from '@theme/Layout';
import SearchMetadata from '@theme/SearchMetadata';
import BlogListPageStructuredData from '@theme/BlogListPage/StructuredData';
import BlogHero from '@site/src/components/blog/BlogHero';
import BlogPostCard from '@site/src/components/blog/BlogPostCard';
import CategoryChips from '@site/src/components/blog/CategoryChips';
import useCategoryFilter from '@site/src/components/blog/useCategoryFilter';
import {CATEGORIES, normalizePost} from '@site/src/components/blog/utils';

// How many cards are revealed at a time by "Load more".
const PAGE_SIZE = 12;

// Cards past the visible window stay in the DOM (hidden with CSS) instead of
// being unmounted, so every post remains a crawlable link in the SSR HTML and
// so a JS-less visitor can be shown the whole list via the <noscript> override
// below. `loading="lazy"` means hidden covers are never downloaded.
const COLLAPSED_CLASS = 'goff-blog-card--collapsed';

const NOSCRIPT_STYLE = `<style>.${COLLAPSED_CLASS}{display:block}</style>`;

function BlogListPageMetadata(props) {
  const {metadata} = props;
  const {
    siteConfig: {title: siteTitle},
  } = useDocusaurusContext();
  const {blogDescription, blogTitle, permalink} = metadata;
  const isBlogOnlyMode = permalink === '/';
  const title = isBlogOnlyMode ? siteTitle : blogTitle;
  return (
    <>
      <PageMetadata title={title} description={blogDescription} />
      <SearchMetadata tag="blog_posts_list" />
    </>
  );
}

BlogListPageMetadata.propTypes = {
  metadata: PropTypes.shape({
    blogDescription: PropTypes.string,
    blogTitle: PropTypes.string,
    permalink: PropTypes.string,
  }).isRequired,
};

function BlogListPageContent({items, metadata: {blogTitle}}) {
  const posts = useMemo(() => items.map(normalizePost), [items]);
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE);

  // Only offer a chip for categories that actually have posts, so a filter can
  // never land on an empty grid.
  const {categories, counts} = useMemo(() => {
    const perCategory = {total: posts.length};
    CATEGORIES.forEach(category => {
      perCategory[category] = posts.filter(
        post => post.category === category
      ).length;
    });
    return {
      categories: CATEGORIES.filter(category => perCategory[category] > 0),
      counts: perCategory,
    };
  }, [posts]);

  const [activeCategory, selectCategory] = useCategoryFilter(categories);

  // The reveal window belongs to the category being shown, so reset it whenever
  // the category changes. Keyed on the resulting category rather than done in
  // the chip handler, so a change coming from the URL (an in-app link to
  // /blog?category=…, back/forward) starts at the same 12 cards a chip does.
  useEffect(() => {
    setVisibleCount(PAGE_SIZE);
  }, [activeCategory]);

  const filtered = activeCategory
    ? posts.filter(post => post.category === activeCategory)
    : posts;

  // The hero is the newest post, and only makes sense on the unfiltered view.
  const hero = activeCategory ? null : filtered[0];
  const gridPosts = hero ? filtered.slice(1) : filtered;
  const remaining = gridPosts.length - visibleCount;

  return (
    <Layout>
      <noscript dangerouslySetInnerHTML={{__html: NOSCRIPT_STYLE}} />
      <main className="mx-auto w-full max-w-[1400px] px-6 py-12">
        {/* The visible page opens on the hero, but the document still needs a
            single top-level heading for assistive tech and crawlers — the hero
            title is an h2 and disappears entirely under an active filter. */}
        <h1 className="sr-only">{blogTitle}</h1>

        {hero && <BlogHero post={hero} />}

        <CategoryChips
          categories={categories}
          active={activeCategory}
          counts={counts}
          onChange={selectCategory}
        />

        {gridPosts.length === 0 && (
          <p className="py-12 text-center text-lg text-[color:var(--goff-main-ff-description)]">
            No article in this category yet.
          </p>
        )}

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {gridPosts.map((post, index) => (
            <div
              key={post.permalink}
              className={index < visibleCount ? undefined : COLLAPSED_CLASS}>
              <BlogPostCard post={post} />
            </div>
          ))}
        </div>

        {remaining > 0 && (
          <div className="mt-12 flex justify-center">
            <button
              type="button"
              onClick={() => setVisibleCount(count => count + PAGE_SIZE)}
              className="cursor-pointer rounded-full border border-solid border-goff-600 bg-transparent px-8 py-3 font-semibold text-goff-700 transition-colors duration-150 hover:bg-goff-600 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-goff-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-goff-400 dark:text-goff-300 dark:hover:bg-goff-500 dark:hover:text-gray-900 dark:focus-visible:ring-offset-[#1b1b1d]">
              Load more ({remaining})
            </button>
          </div>
        )}
      </main>
    </Layout>
  );
}

BlogListPageContent.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({content: PropTypes.any.isRequired}))
    .isRequired,
  metadata: PropTypes.shape({blogTitle: PropTypes.string}).isRequired,
};

export default function BlogListPage(props) {
  return (
    <HtmlClassNameProvider
      className={clsx(
        ThemeClassNames.wrapper.blogPages,
        ThemeClassNames.page.blogListPage
      )}>
      <BlogListPageMetadata {...props} />
      <BlogListPageStructuredData {...props} />
      <BlogListPageContent {...props} />
    </HtmlClassNameProvider>
  );
}

import React from 'react';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CategoryBadge from '../CategoryBadge';
import PostMeta from '../PostMeta';
import {postShape} from '../propTypes';

// Featured (latest) post. Two columns from `lg` up, stacked below.
export default function BlogHero({post}) {
  const image = useBaseUrl(post.image);
  return (
    <Link
      to={post.permalink}
      className="group mb-12 grid grid-cols-1 gap-8 overflow-hidden rounded-3xl border border-solid border-gray-200 bg-white no-underline shadow-sm transition-shadow duration-200 hover:no-underline hover:shadow-lg dark:border-gray-700 dark:bg-[#1f2024] lg:grid-cols-2 lg:gap-0">
      {/* See BlogPostCard: `object-contain` keeps wide banners readable. */}
      <div className="aspect-[1200/630] overflow-hidden bg-gray-50 dark:bg-gray-800 lg:aspect-auto lg:h-full">
        <img
          src={image}
          alt={post.imageAlt}
          width={1200}
          height={630}
          // The camelCase `fetchPriority` the lint rule wants is React 19+.
          // react-dom 18.3.1 does not know that prop and warns on it, so emit
          // the plain DOM attribute instead.
          // eslint-disable-next-line react/no-unknown-property
          fetchpriority="high"
          className="h-full w-full object-contain transition-transform duration-300 group-hover:scale-105"
        />
      </div>
      <div className="flex flex-col gap-4 p-8 lg:justify-center lg:p-12">
        <div className="flex flex-wrap items-center gap-3">
          <span className="text-xs font-semibold uppercase tracking-wide text-[color:var(--goff-main-ff-description)]">
            Latest article
          </span>
          <CategoryBadge category={post.category} />
        </div>
        <h2 className="m-0 text-2xl font-bold leading-tight text-gray-800 dark:text-gray-50 sm:text-3xl lg:text-4xl">
          {post.title}
        </h2>
        {post.description && (
          <p className="m-0 line-clamp-4 text-lg leading-relaxed text-[color:var(--goff-main-ff-description)]">
            {post.description}
          </p>
        )}
        <PostMeta
          authors={post.authors}
          formattedDate={post.formattedDate}
          readingTime={post.readingTime}
          size="large"
        />
        {/* Not --ifm-color-primary-dark: in this palette that resolves to a
            light mint (#73c6b6) that fails contrast on the white card. */}
        <span className="inline-flex items-center gap-1 font-semibold text-goff-700 dark:text-goff-300">
          Read article <span aria-hidden="true">→</span>
        </span>
      </div>
    </Link>
  );
}

BlogHero.propTypes = {
  post: postShape.isRequired,
};

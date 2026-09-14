import React from 'react';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CategoryBadge from '../CategoryBadge';
import PostMeta from '../PostMeta';
import {postShape} from '../propTypes';

// The whole card is a single link: one tab stop per post, and the entire
// surface is clickable.
const CARD_CLASS =
  'group flex h-full flex-col overflow-hidden rounded-2xl border border-solid border-gray-200 bg-white no-underline shadow-sm transition-shadow duration-200 hover:no-underline hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-goff-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:border-gray-700 dark:bg-[#1f2024] dark:focus-visible:ring-offset-[#1b1b1d]';

export default function BlogPostCard({post}) {
  const image = useBaseUrl(post.image);
  return (
    <Link to={post.permalink} className={CARD_CLASS}>
      {/* Fixed aspect ratio + intrinsic size so the grid never reflows while
          covers load (CLS). `object-contain` rather than `object-cover`: our
          covers range from 1.8:1 to 3.8:1, and cropping a wide banner to the
          card ratio cuts the words off it. */}
      <div className="aspect-[1200/630] overflow-hidden bg-gray-50 dark:bg-gray-800">
        <img
          src={image}
          alt={post.imageAlt}
          width={1200}
          height={630}
          loading="lazy"
          className="h-full w-full object-contain transition-transform duration-300 group-hover:scale-105"
        />
      </div>
      <div className="flex flex-1 flex-col gap-3 p-6">
        <CategoryBadge category={post.category} />
        <h3 className="m-0 line-clamp-2 text-xl font-bold leading-snug text-gray-800 dark:text-gray-50">
          {post.title}
        </h3>
        {post.description && (
          <p className="m-0 line-clamp-3 leading-relaxed text-[color:var(--goff-main-ff-description)]">
            {post.description}
          </p>
        )}
        <div className="mt-auto pt-3">
          <PostMeta
            authors={post.authors}
            formattedDate={post.formattedDate}
            readingTime={post.readingTime}
          />
        </div>
      </div>
    </Link>
  );
}

BlogPostCard.propTypes = {
  post: postShape.isRequired,
};

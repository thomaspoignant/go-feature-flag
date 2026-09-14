# Writing a blog post

> This file is excluded from the blog build (see `exclude` in
> `website/docusaurus.config.js`), so it is never published as a post.

`/blog` renders every post as a card in a grid, with a hero for the latest post
and one filter chip per category. That layout only works if every post carries a
cover image and a category, so three front matter fields are **required** and
checked in CI.

## Create the post

```
website/blog/YYYY-MM-DD-my-post-slug/
├── index.md      (or index.mdx if you need JSX)
└── cover.png     the cover image, next to the post
```

Front matter:

```yaml
---
title: 'A title that reads well at 380px wide in a card'
description: Two lines max. This is the card excerpt and the meta description.
authors: [thomaspoignant]
tags: [openfeature, exporter]
image: ./cover.png
image_alt: 'A short description of what the cover shows'
category: Product
---
```

Add `<!-- truncate -->` after the first paragraph or two, as usual.

## Required fields

### `image`

The cover shown on the card and the hero. Docusaurus also reuses it as
`og:image`, so this is the picture people see when the post is shared.

- **Relative path** (`./cover.png`) — preferred. The file lives next to
  `index.md` and goes through the asset pipeline, so the URL can never go stale.
- **Absolute path** (`/img/...`) — for images shared across posts, resolved
  under `website/static/`.
- **Remote URLs are rejected.** They break as soon as the remote changes.

Spec: **1200 px wide** (16:9-ish), readable at ~380 px wide, and visually
consistent with the rest of the blog. If you have nothing better yet, use
`/img/logo/x-card.png` — but a post with its own cover always looks better in
the grid.

**Under 200 KB — this one is enforced, `npm run lint:blog` fails above it.**
Every cover is downloaded by everyone who opens `/blog`, not just readers of
your post, so the budget is not advisory. To hit it without extra tooling:

```bash
# resize to 1200 px wide and re-encode (macOS, built in)
sips -Z 1200 -s format jpeg -s formatOptions 78 banner.png --out banner.jpg
```

Keep PNG for screenshots and flat-colour graphics if it fits the budget after
resizing; use JPEG for photos and detailed illustrations, where PNG will not.

### `image_alt`

Alt text for the cover. Describe what the image _shows_ ("The Grafana and GO
Feature Flag logos side by side"), not what the article is about — the title
already says that.

### `category`

Exactly one of:

| Category      | Use it for                                                   |
| ------------- | ------------------------------------------------------------ |
| `Product`     | Releases, new capabilities, provider and relay proxy changes |
| `Engineering` | Tutorials, deep dives, best practices, comparisons           |
| `Community`   | Talks, conferences, podcasts, adoption news, governance      |
| `Case study`  | A named user telling their story                             |

Categories are deliberately few — they are the filter chips on `/blog`. Tags
stay free-form and are used for the tag pages, not the chips.

Adding a category means editing three places: `CATEGORIES` in
`website/src/components/blog/utils.js`, `CATEGORIES` in
`website/scripts/check-blog-frontmatter.mjs`, and the table above.

## Check your work

```bash
cd website
npm run lint:blog   # covers + categories (this is what CI runs)
npm start           # then open http://localhost:3000/blog
```

On `/blog`, confirm your card renders with the cover, the right chip filters it
in, and the title is not clipped mid-word.

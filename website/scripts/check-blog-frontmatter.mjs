#!/usr/bin/env node
/**
 * Blog front matter check.
 *
 * The blog index is a card grid, so a post without a cover leaves a hole in it.
 * This guards the fields the grid depends on:
 *
 *   image      - cover, also reused by Docusaurus as og:image
 *   image_alt  - alt text for that cover
 *   category   - one of the values rendered as a filter chip
 *
 * Run with `npm run lint:blog` from the website/ directory.
 */
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import YAML from 'yaml';

const WEBSITE_DIR = path.resolve(fileURLToPath(import.meta.url), '../..');
const BLOG_DIR = path.join(WEBSITE_DIR, 'blog');
const STATIC_DIR = path.join(WEBSITE_DIR, 'static');

// Keep in sync with CATEGORIES in src/components/blog/utils.js.
const CATEGORIES = ['Product', 'Engineering', 'Community', 'Case study'];

// Hard budget. The blog index shows the cover of every post, so an oversized
// one is paid for by everyone landing on /blog, not just readers of that post.
// All current covers are under this; see website/blog/README.md for how to
// shrink a new one.
const COVER_SIZE_MAX_BYTES = 200 * 1024;

// Mirrors the `exclude` list in docusaurus.config.js: files Docusaurus does not
// publish as posts must not be checked as posts either.
const isExcluded = name => name.startsWith('_') || name === 'README.md';

/** Walks the blog tree the way Docusaurus does, rather than one level deep. */
const listPostFiles = (dir = BLOG_DIR) =>
  fs
    .readdirSync(dir, {withFileTypes: true})
    .flatMap(entry => {
      if (isExcluded(entry.name)) {
        return [];
      }
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        return listPostFiles(full);
      }
      return /\.mdx?$/.test(entry.name) ? [full] : [];
    })
    .sort();

/**
 * Reads the leading `---` block only. The regex has no `m` flag on purpose, so
 * it anchors at byte 0 and cannot match an `image:`-looking line inside the
 * post body — `2023-02-20-lint-your-feature-flags` has a literal `image: ubuntu`
 * at column 0 in a fenced CI config block.
 *
 * Returns null for a file with no front matter, which is reported as a normal
 * error rather than crashing the run.
 */
const parseFrontMatter = raw => {
  const match = /^---\r?\n([\s\S]*?)\r?\n---/.exec(raw);
  if (!match) {
    return null;
  }
  try {
    return YAML.parse(match[1]) ?? {};
  } catch {
    return null;
  }
};

/** Resolves a front matter `image` value to a path on disk, or null if remote. */
const resolveCover = (postFile, image) => {
  if (/^https?:\/\//.test(image)) {
    return null;
  }
  if (image.startsWith('/')) {
    return path.join(STATIC_DIR, image);
  }
  return path.resolve(path.dirname(postFile), image);
};

const errors = [];
const files = listPostFiles();

for (const file of files) {
  const rel = path.relative(WEBSITE_DIR, file);
  const fail = message => errors.push(`${rel}: ${message}`);
  const fm = parseFrontMatter(fs.readFileSync(file, 'utf8'));

  if (fm === null) {
    fail('no readable YAML front matter block');
    continue;
  }

  if (typeof fm.image !== 'string' || fm.image.trim() === '') {
    fail(
      'missing `image:` front matter — every post needs a cover for the blog card grid'
    );
  } else {
    const cover = resolveCover(file, fm.image);
    if (cover === null) {
      fail(
        `\`image: ${fm.image}\` is a remote URL — use a path relative to the post (./cover.png) or a /img/... path under static/`
      );
    } else if (!fs.existsSync(cover)) {
      fail(`\`image: ${fm.image}\` does not exist (looked for ${cover})`);
    } else {
      const {size} = fs.statSync(cover);
      if (size > COVER_SIZE_MAX_BYTES) {
        fail(
          `cover is ${Math.round(size / 1024)} KB — the budget is ${
            COVER_SIZE_MAX_BYTES / 1024
          } KB, resize it to 1200 px wide (see website/blog/README.md)`
        );
      }
    }
  }

  if (typeof fm.image_alt !== 'string' || fm.image_alt.trim() === '') {
    fail('missing `image_alt:` front matter — describe the cover image');
  }

  if (typeof fm.category !== 'string' || fm.category.trim() === '') {
    fail(
      `missing \`category:\` front matter — one of ${CATEGORIES.join(', ')}`
    );
  } else if (!CATEGORIES.includes(fm.category)) {
    fail(
      `unknown \`category: ${fm.category}\` — must be one of ${CATEGORIES.join(', ')}`
    );
  }
}

if (errors.length > 0) {
  errors.forEach(error => console.error(`❌ ${error}`));
  console.error(
    `\n${errors.length} problem(s) in ${files.length} blog post(s). See website/blog/README.md.`
  );
  process.exit(1);
}

console.log(`✅ ${files.length} blog post(s) have a valid cover and category.`);

import {useCallback, useEffect, useState} from 'react';
import {useHistory, useLocation} from '@docusaurus/router';
import {slugify} from './utils';

export const CATEGORY_PARAM = 'category';

/**
 * Category filter state, mirrored into the `?category=` query param so a
 * filtered view can be shared or bookmarked.
 *
 * The value is read in an effect rather than during render: the server renders
 * `/blog` with no query string, so reading it on the first client render would
 * produce a hydration mismatch. The unfiltered list paints first, then the
 * filter from the URL is applied.
 *
 * The effect depends on `location.search`, which makes the URL the single
 * source of truth — an in-app link to `/blog?category=product` re-filters
 * without a remount.
 *
 * We use `history.replace` so repeatedly clicking chips does not fill the back
 * stack — "back" leaves /blog rather than stepping through past selections.
 *
 * @param {string[]} available categories that actually have posts. A param that
 *   is not one of them (stale link, category emptied out) falls back to the
 *   unfiltered list rather than rendering an empty grid with no chip selected.
 */
export default function useCategoryFilter(available) {
  const history = useHistory();
  const {search} = useLocation();
  const [active, setActive] = useState(null);

  useEffect(() => {
    const param = new URLSearchParams(search).get(CATEGORY_PARAM);
    const match = available.find(category => slugify(category) === param);
    setActive(match ?? null);
  }, [search, available]);

  const select = useCallback(
    category => {
      setActive(category);
      const params = new URLSearchParams(window.location.search);
      if (category) {
        params.set(CATEGORY_PARAM, slugify(category));
      } else {
        params.delete(CATEGORY_PARAM);
      }
      const query = params.toString();
      history.replace(
        `${window.location.pathname}${query ? `?${query}` : ''}${
          window.location.hash
        }`
      );
    },
    [history]
  );

  return [active, select];
}

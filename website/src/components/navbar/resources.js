import {
  FaBlog,
  FaBook,
  FaStar,
  FaLaptopCode,
  FaBalanceScale,
} from 'react-icons/fa';
import {
  generateMenuColumn,
  generateFeaturedCard,
  generateMegaMenu,
} from './menu.js';

// Generate the HTML string for the "Resources" mega-menu (3 columns, the last
// being a GrowthBook-style featured "Explore" card).
export function generateResourcesDropdownHTML() {
  const columns = [
    {
      title: 'Resources',
      items: [
        {label: 'Blog', href: '/blog', icon: FaBlog},
        {label: 'Documentation', href: '/docs', icon: FaBook},
        {
          label: 'Best Practices',
          href: '/blog/feature-flag-best-practice',
          icon: FaStar,
        },
      ],
    },
    {
      title: 'Compare',
      items: [
        {
          label: 'Why GO Feature Flag',
          href: '/product/why_go_feature_flag',
          icon: FaLaptopCode,
        },
        {
          label: 'GO Feature Flag vs Others',
          href: '/blog/best-opensource-feature-flag-tools',
          icon: FaBalanceScale,
        },
      ],
    },
  ];

  const columnsHtml = columns
    .map((column, index) =>
      generateMenuColumn({...column, withDivider: index > 0})
    )
    .join('');

  /*const featuredHtml = generateFeaturedCard({
    columnTitle: 'Explore',
    title: 'How to migrate <br />from Launch Darkly <br />to GO Feature Flag',
    href: '/resources/migrate-from-launchdarkly',
  });*/

  return generateMegaMenu(columnsHtml /*+ featuredHtml*/);
}

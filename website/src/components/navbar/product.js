import {
  FaFlag,
  FaToggleOn,
  FaVial,
  FaRobot,
  FaCode,
  FaPuzzlePiece,
  FaDatabase,
} from 'react-icons/fa';
import {MdRocketLaunch} from 'react-icons/md';
import {RiOpenSourceFill} from 'react-icons/ri';
import {generateMenuColumn, generateMegaMenu} from './menu.js';

// Generate the HTML string for the "Product" mega-menu (2 columns).
export function generateProductDropdownHTML() {
  const columns = [
    {
      title: 'Product',
      items: [
        {
          label: 'Feature Flag',
          href: '/product/what-are-feature-flags',
          icon: FaFlag,
        },
        {
          label: 'OpenFeature',
          href: '/product/open_feature_support',
          icon: FaToggleOn,
        },
        {label: 'Rollouts', href: '/product/rollouts', icon: MdRocketLaunch},
        {
          label: 'Test in Production',
          href: '/product/test_in_production',
          icon: FaVial,
        },
        {label: 'AI', href: '/product/ai', icon: FaRobot},
      ],
    },
    {
      title: 'Platform',
      items: [
        {
          label: 'OpenSource',
          href: '/product/open_source',
          icon: RiOpenSourceFill,
        },
        {label: 'SDKs', href: '/docs/sdk', icon: FaCode},
        {
          label: 'Integrations',
          href: '/product/integrations',
          icon: FaPuzzlePiece,
        },
        {label: 'Data', href: '/product/data', icon: FaDatabase},
      ],
    },
  ];

  const columnsHtml = columns
    .map((column, index) =>
      generateMenuColumn({...column, withDivider: index > 0})
    )
    .join('');

  return generateMegaMenu(columnsHtml);
}

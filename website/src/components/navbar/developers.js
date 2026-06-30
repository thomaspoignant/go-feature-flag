import {FaBook, FaCode, FaPencilAlt, FaGithub, FaSlack} from 'react-icons/fa';
import {LuBug} from 'react-icons/lu';
import {IoChatboxEllipsesOutline} from 'react-icons/io5';
import {
  generateMenuColumn,
  generateHelpSection,
  generateMegaMenu,
} from './menu.js';

// Generate the HTML string for the "Developers" mega-menu: two columns plus a
// "Get Help" footer row of cards.
export function generateDevelopersDropdownHTML() {
  const columns = [
    {
      title: 'Developers',
      items: [
        {label: 'Documentation', href: '/docs', icon: FaBook},
        {label: 'SDKs', href: '/docs/sdk', icon: FaCode},
        {label: 'Flag Editor', href: '/editor', icon: FaPencilAlt},
      ],
    },
    {
      title: 'Resources',
      items: [
        {
          label: 'GitHub',
          href: 'https://github.com/thomaspoignant/go-feature-flag',
          icon: FaGithub,
          external: true,
        },
        {
          label: 'Release Notes',
          href: 'https://github.com/thomaspoignant/go-feature-flag/releases',
          icon: FaGithub,
          external: true,
        },
        {label: 'Community', href: '/slack', icon: FaSlack},
      ],
    },
  ];

  const columnsHtml = columns
    .map((column, index) =>
      generateMenuColumn({...column, withDivider: index > 0})
    )
    .join('');

  const helpHtml = generateHelpSection({
    title: 'Get Help',
    cards: [
      {
        title: 'Talk to the community',
        description:
          'Connect with GO Feature Flag users and ask your questions on Slack.',
        href: '/slack',
        icon: IoChatboxEllipsesOutline,
      },
      {
        title: 'Found a bug? Have an idea?',
        description: 'Open an issue on GitHub to help improve GO Feature Flag.',
        href: 'https://github.com/thomaspoignant/go-feature-flag/issues/new/choose',
        icon: LuBug,
        external: true,
      },
    ],
  });

  return generateMegaMenu(columnsHtml, helpHtml);
}

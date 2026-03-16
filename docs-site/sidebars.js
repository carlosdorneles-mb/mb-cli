// @ts-check

/**
 * Sidebar da documentação do MB CLI.
 * @type {import('@docusaurus/plugin-content-docs').SidebarsConfig}
 */
const sidebars = {
  docsSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Guia',
      collapsed: false,
      items: [
        'getting-started',
        'creating-plugins',
        'plugin-commands',
        'global-flags',
        'environment-variables',
        'security',
      ],
    },
    {
      type: 'category',
      label: 'Referência técnica',
      collapsed: true,
      items: [
        'architecture',
        'plugins',
        'helpers-shell',
        'reference',
      ],
    },
  ],
};

export default sidebars;

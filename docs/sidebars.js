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
        'guide/getting-started',
        'guide/global-flags',
        'guide/plugin-commands',
        'guide/creating-plugins',
        'guide/environment-variables',
        'guide/security',
      ],
    },
    {
      type: 'category',
      label: 'Referência técnica',
      collapsed: true,
      items: [
        'technical-reference/architecture',
        'technical-reference/plugins',
        'technical-reference/helpers-shell',
        'technical-reference/reference',
        'technical-reference/versioning-and-release',
      ],
    },
  ],
};

export default sidebars;

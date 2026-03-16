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
        'comandos-plugins',
        'flags-globais',
        'variaveis-ambiente',
      ],
    },
    {
      type: 'category',
      label: 'Referência técnica',
      collapsed: true,
      items: [
        'arquitetura',
        'plugins',
        'reference',
      ],
    },
  ],
};

export default sidebars;

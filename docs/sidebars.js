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
      label: 'Começar',
      collapsed: false,
      items: [
        'getting-started/index',
        'getting-started/local-development',
      ],
    },
    {
      type: 'category',
      label: 'Guia do Usuário',
      collapsed: false,
      items: [
        'user-guide/environment-variables',
        'user-guide/plugin-commands',
        'user-guide/global-flags',
        'user-guide/security',
      ],
    },
    {
      type: 'category',
      label: 'Criar Plugins',
      collapsed: true,
      items: [
        'plugin-authoring/create-a-plugin',
        'plugin-authoring/shell-helpers',
      ],
    },
    {
      type: 'category',
      label: 'Referência técnica',
      collapsed: true,
      link: { type: 'generated-index' },
      items: [
        'technical-reference/architecture',
        'technical-reference/plugins',
        'technical-reference/plugin-invocation-context',
        'technical-reference/cli-config',
        'technical-reference/reference',
        'technical-reference/versioning-and-release',
      ],
    },
  ],
};

export default sidebars;

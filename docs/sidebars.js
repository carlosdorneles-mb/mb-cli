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
      collapsed: true,
      link: { type: 'doc', id: 'getting-started/index' },
      items: [
        'getting-started/index',
        'getting-started/installation',
        'getting-started/local-development',
      ],
    },
    {
      type: 'category',
      label: 'Guia do Usuário',
      collapsed: true,
      link: { type: 'generated-index' },
      items: [
        'user-guide/environment-variables',
        'user-guide/mbcli-yaml',
        'user-guide/plugin-commands',
        'user-guide/global-flags',
        'user-guide/security',
      ],
    },
    {
      type: 'category',
      label: 'Comandos do CLI',
      collapsed: true,
      link: { type: 'generated-index' },
      items: [
        'commands/envs',
        'commands/alias',
        'commands/plugins',
        'commands/run',
        'commands/update',
        'commands/completion',
        'commands/help',
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
        'technical-reference/versioning-and-release',
      ],
    },
    {
      type: 'category',
      label: 'Criar Plugins',
      collapsed: true,
      link: { type: 'generated-index' },
      items: [
        'plugin-authoring/create-a-plugin',
        'plugin-authoring/shell-helpers',
      ],
    },
  ],
};

export default sidebars;

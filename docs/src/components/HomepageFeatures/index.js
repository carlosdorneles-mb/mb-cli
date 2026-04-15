import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    icon: '⚡',
    title: 'Comandos dinâmicos',
    description: (
      <>
        Plugins viram comandos <code>mb &lt;categoria&gt; &lt;comando&gt;</code> automaticamente.
        Instale por URL Git ou registre um path local para desenvolvimento.
      </>
    ),
    link: '/docs/user-guide/plugin-commands',
    linkText: 'Comandos de plugins',
  },
  // {
  //   icon: '🗄️',
  //   title: 'Cache e sync',
  //   description: (
  //     <>
  //       Cache SQLite guarda plugins e categorias; <code>mb plugins sync</code> atualiza
  //       a partir do diretório de plugins e dos paths locais registrados.
  //     </>
  //   ),
  //   link: '/docs/technical-reference/architecture',
  //   linkText: 'Arquitetura',
  // },
  {
    icon: '🔒',
    title: 'Ambiente e plugins',
    description: (
      <>
        Variáveis mescladas (sistema, defaults, <code>--env</code>) e injetadas só no processo do plugin.
        Crie plugins com <code>manifest.yaml</code> e scripts ou binários.
      </>
    ),
    links: [
      { to: '/docs/user-guide/environment-variables', text: 'Variáveis de ambiente' },
      { to: '/docs/plugin-authoring/create-a-plugin', text: 'Criar um plugin' },
    ],
  },
  {
    icon: '⌨️',
    title: 'Atalhos pessoais',
    description: (
      <>
        Defina nomes curtos para comandos que você repete; o perfil do shell carrega o atalho. Com{' '}
        <code>mb run &lt;nome&gt;</code> o mesmo alias usa o ambiente mesclado do MB (incluindo{' '}
        <code>--vault</code> opcional no <code>mb alias set</code>).
      </>
    ),
    link: '/docs/commands/alias',
    linkText: 'Referência mb alias',
  },
];

function Feature({icon, title, description, link, linkText, links}) {
  return (
    <div className={clsx('col col--4')}>
      <div className={styles.featureCard}>
        <div className={styles.featureIcon}>{icon}</div>
        <Heading as="h3" className={styles.featureTitle}>
          {title}
        </Heading>
        <div className={styles.featureDescription}>{description}</div>
        <div className={styles.featureLinks}>
          {link && (
            <Link to={link} className={styles.featureLink}>
              {linkText} →
            </Link>
          )}
          {links &&
            links.map((l, i) => (
              <span key={i} className={styles.featureLinksRow}>
                <Link to={l.to} className={styles.featureLink}>
                  {l.text} →
                </Link>
              </span>
            ))}
        </div>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <Heading as="h2" className={styles.featuresTitle}>
          Como funciona
        </Heading>
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

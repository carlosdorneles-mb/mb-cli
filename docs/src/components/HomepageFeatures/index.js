import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Comandos dinâmicos',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        Plugins viram comandos <code>mb &lt;categoria&gt; &lt;comando&gt;</code> automaticamente.
        Instale por URL Git ou registre um path local para desenvolvimento.
        <br />
        <Link to="/docs/user-guide/plugin-commands">Comandos de plugins</Link>
      </>
    ),
  },
  {
    title: 'Cache e sync',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        Cache SQLite guarda plugins e categorias; <code>mb plugins sync</code> atualiza
        a partir do diretório de plugins e dos paths locais registrados.
        <br />
        <Link to="/docs/technical-reference/architecture">Arquitetura</Link>
      </>
    ),
  },
  {
    title: 'Ambiente e plugins',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        Variáveis mescladas (sistema, defaults, <code>--env</code>) e injetadas só no processo do plugin.
        Crie plugins com <code>manifest.yaml</code> e scripts ou binários.
        <br />
        <Link to="/docs/user-guide/environment-variables">Variáveis de ambiente</Link>
        {' · '}
        <Link to="/docs/plugin-authoring/create-a-plugin">Criar um plugin</Link>
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

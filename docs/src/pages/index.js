import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';
import styles from './index.module.css';
import {useEffect, useState} from 'react';

// Animação de digitação: exibe o texto caractere por caractere
function TypingAnimation({text, delay = 0, speed = 60}) {
  const [displayed, setDisplayed] = useState('');
  const [started, setStarted] = useState(false);

  useEffect(() => {
    const startTimeout = setTimeout(() => {
      setStarted(true);
    }, delay);
    return () => clearTimeout(startTimeout);
  }, [delay]);

  useEffect(() => {
    if (!started) return;
    let index = 0;
    const interval = setInterval(() => {
      if (index <= text.length) {
        setDisplayed(text.slice(0, index));
        index++;
      } else {
        clearInterval(interval);
      }
    }, speed);
    return () => clearInterval(interval);
  }, [text, speed, started]);

  return <>{displayed}</>;
}

// Output que aparece após um delay
function TerminalOutput({children, delay = 0}) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setVisible(true);
    }, delay);
    return () => clearTimeout(timeout);
  }, [delay]);

  if (!visible) return null;
  return <>{children}</>;
}

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <div className={styles.heroContent}>
          <div className={styles.heroText}>
            <Heading as="h1" className={styles.heroTitle}>
              Orquestre plugins com{' '}
              <span className={styles.heroGradientText}>segurança</span> e{' '}
              <span className={styles.heroGradientText}>performance</span>
            </Heading>
            <p className={styles.heroSubtitle}>
              MB CLI é uma ferramenta CLI em Go que transforma plugins em comandos dinâmicos,
              com cache, injeção segura de variáveis de ambiente e helpers de shell poderosos.
            </p>
            <div className={styles.heroButtons}>
              <Link
                className={clsx('button', styles.buttonPrimary)}
                to="/docs/intro">
                Começar agora
              </Link>
              <Link
                className={clsx('button', styles.buttonSecondary)}
                to="/docs/getting-started/">
                Guia rápido
              </Link>
            </div>
          </div>
          <div className={styles.heroTerminal}>
            <div className={styles.terminalWindow}>
              <div className={styles.terminalHeader}>
                <div className={styles.terminalDotRed} />
                <div className={styles.terminalDotYellow} />
                <div className={styles.terminalDotGreen} />
                <span className={styles.terminalTitle}>terminal</span>
              </div>
              <div className={styles.terminalBody}>
                {/* Linha 1: digitando mb plugins sync */}
                <div className={styles.terminalLine}>
                  <span className={styles.terminalPrompt}>$</span>{' '}
                  <span className={styles.terminalCommand}>
                    <TypingAnimation text="mb plugins sync" delay={500} speed={60} />
                  </span>
                </div>

                {/* Outputs do sync */}
                <TerminalOutput delay={1800}>
                <div className={clsx(styles.terminalOutput, styles.terminalInfo)}>
                    <span className={styles.terminalInfoLabel}>INFO</span>{' '}Comando "tools" adicionado
                  </div>
                </TerminalOutput>
                <TerminalOutput delay={2300}>
                <div className={clsx(styles.terminalOutput, styles.terminalInfo)}>
                    <span className={styles.terminalInfoLabel}>INFO</span>{' '}Comando "tools/vscode" adicionado
                  </div>
                </TerminalOutput>
                <TerminalOutput delay={2800}>
                <div className={clsx(styles.terminalOutput, styles.terminalInfo)}>
                    <span className={styles.terminalInfoLabel}>INFO</span>{' '}Comando "tools/postman" adicionado
                  </div>
                </TerminalOutput>
                <TerminalOutput delay={3300}>
                  <div className={clsx(styles.terminalOutput, styles.terminalInfo)}>
                    <span className={styles.terminalInfoLabel}>INFO</span>{' '}Comando "tools/podman" adicionado
                  </div>
                </TerminalOutput>

                {/* Linha 2: digitando mb tools deploy */}
                <TerminalOutput delay={4000}>
                  <div className={styles.terminalLine}>
                    <span className={styles.terminalPrompt}>$</span>{' '}
                    <span className={styles.terminalCommand}>
                      <TypingAnimation text="mb tools deploy --env production" speed={50} />
                    </span>
                  </div>
                </TerminalOutput>

                {/* Output do deploy */}
                <TerminalOutput delay={5800}>
                  <div className={clsx(styles.terminalOutput, styles.terminalInfo)}>
                    <span className={styles.terminalInfoLabel}>INFO</span>{' '}Deploy concluído com sucesso
                  </div>
                </TerminalOutput>

                {/* Cursor piscando no final */}
                <TerminalOutput delay={6100}>
                  <div className={styles.terminalCursor}>▊</div>
                </TerminalOutput>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

function QuickStartSection() {
  const quickStartItems = [
    {
      icon: '📦',
      title: 'Instalar',
      description: 'Adicione plugins por URL Git ou path local',
      code: 'mb plugins add https://github.com/org/repo',
    },
    {
      icon: '🔄',
      title: 'Sincronizar',
      description: 'Cache SQLite atualiza automaticamente',
      code: 'mb plugins sync',
    },
    {
      icon: '🚀',
      title: 'Executar',
      description: 'Comandos dinâmicos prontos para uso',
      code: 'mb <categoria> <comando> [flags]',
    },
    {
      icon: '🛠️',
      title: 'Criar plugin',
      description: 'manifest.yaml + script = comando',
      code: 'mb plugins add ./meu-plugin --package meu-plugin',
    },
  ];

  return (
    <section className={styles.quickStartSection}>
      <div className="container">
        <Heading as="h2" className={styles.sectionTitle}>
          Rápido e simples
        </Heading>
        <p className={styles.sectionSubtitle}>
          Quatro passos para começar a usar plugins no seu projeto
        </p>
        <div className={styles.quickStartGrid}>
          {quickStartItems.map((item, idx) => (
            <div key={idx} className={styles.quickStartCard}>
              <div className={styles.quickStartIcon}>{item.icon}</div>
              <Heading as="h3" className={styles.quickStartTitle}>
                {item.title}
              </Heading>
              <p className={styles.quickStartDescription}>{item.description}</p>
              <code className={styles.quickStartCode}>{item.code}</code>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function KeyFeaturesSection() {
  const keyFeatures = [
    {
      icon: '⚡',
      title: 'Comandos dinâmicos',
      description:
        'Plugins viram comandos automaticamente na árvore CLI. Sem boilerplate, sem configuração complexa.',
      link: '/docs/user-guide/plugin-commands',
      linkText: 'Ver comandes',
    },
    {
      icon: '🔒',
      title: 'Ambiente seguro',
      description:
        'Variáveis mescladas (sistema, defaults, --env) injetadas só no processo do plugin com controle total.',
      link: '/docs/user-guide/environment-variables',
      linkText: 'Variáveis de ambiente',
    },
    {
      icon: '⌨️',
      title: 'Atalhos pessoais',
      description:
        'Defina nomes curtos para comandos repetidos; o perfil carrega o atalho no shell. Com mb run <nome>, o mesmo alias usa o ambiente mesclado do MB (vault opcional).',
      link: '/docs/commands/alias',
      linkText: 'Referência mb alias',
    },
    // {
    //   icon: '🗄️',
    //   title: 'Cache inteligente',
    //   description:
    //     'SQLite guarda plugins, categorias e hashes. Sync detecta mudanças e atualiza apenas o necessário.',
    //   link: '/docs/technical-reference/architecture',
    //   linkText: 'Arquitetura',
    // },
    {
      icon: '🛠️',
      title: 'Helpers de shell',
      description:
        'Biblioteca de helpers para logs, memória, Kubernetes, Homebrew, Flatpak e muito mais.',
      link: '/docs/plugin-authoring/shell-helpers',
      linkText: 'Ver helpers',
    },
    {
      icon: '📝',
      title: 'Manifest YAML',
      description:
        'Defina comandos, flags, entrypoints e grupos de help com manifests simples em YAML.',
      link: '/docs/plugin-authoring/create-a-plugin',
      linkText: 'Criar um plugin',
    },
    {
      icon: '🌐',
      title: 'Remoto ou local',
      description:
        'Instale plugins de repositórios Git (com tags) ou registre paths locais para desenvolvimento.',
      link: '/docs/getting-started/',
      linkText: 'Começar',
    },
  ];

  return (
    <section className={styles.keyFeaturesSection}>
      <div className="container">
        <Heading as="h2" className={styles.sectionTitle}>
          Recursos poderosos
        </Heading>
        <p className={styles.sectionSubtitle}>
          Tudo que você precisa para gerenciar e executar plugins com confiança
        </p>
        <div className={styles.keyFeaturesGrid}>
          {keyFeatures.map((feature, idx) => (
            <div key={idx} className={styles.keyFeatureCard}>
              <div className={styles.keyFeatureIcon}>{feature.icon}</div>
              <Heading as="h3" className={styles.keyFeatureTitle}>
                {feature.title}
              </Heading>
              <p className={styles.keyFeatureDescription}>{feature.description}</p>
              <Link to={feature.link} className={styles.keyFeatureLink}>
                {feature.linkText} →
              </Link>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function InstallSection() {
  const [os, setOs] = useState('macos');
  const [copied, setCopied] = useState(false);

  const installCommands = {
    macos: 'curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash',
    linux: 'curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash',
  };

  const command = installCommands[os];

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      const textArea = document.createElement('textarea');
      textArea.value = command;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  useEffect(() => {
    const userAgent = navigator.userAgent;
    if (userAgent.includes('Linux')) {
      setOs('linux');
    } else if (userAgent.includes('Mac') || userAgent.includes('macOS')) {
      setOs('macos');
    }
  }, []);

  return (
    <section className={styles.installSection}>
      <div className="container">
        <Heading as="h2" className={styles.installTitle}>
          Instale o <span className={styles.heroGradientText}>MB CLI</span>
        </Heading>
        <p className={styles.installSubtitle}>
          Disponível para Linux e macOS. Execute o comando abaixo no seu terminal:
        </p>
        <div className={styles.installOsTabs}>
          <button
            className={clsx(styles.osTab, os === 'macos' && styles.osTabActive)}
            onClick={() => setOs('macos')}>
            macOS
          </button>
          <button
            className={clsx(styles.osTab, os === 'linux' && styles.osTabActive)}
            onClick={() => setOs('linux')}>
            Linux
          </button>
        </div>
        <div className={styles.installCodeBlock}>
          <span className={styles.installCodePrompt}>{'>'}_</span>
          <code className={styles.installCode}>{command}</code>
          <button className={styles.copyButton} onClick={handleCopy} title="Copiar comando">
            {copied ? (
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            ) : (
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
            )}
          </button>
        </div>
      </div>
    </section>
  );
}

function CallToActionSection() {
  return (
    <section className={styles.ctaSection}>
      <div className={clsx('container', styles.ctaContainer)}>
        <Heading as="h2" className={styles.ctaTitle}>
          Pronto para começar?
        </Heading>
        <p className={styles.ctaSubtitle}>
          Instale o MB CLI e comece a usar plugins no seu projeto agora
        </p>
        <div className={styles.ctaButtons}>
          <Link
            className={clsx('button', styles.buttonPrimary)}
            to="/docs/intro">
            Ver documentação
          </Link>
          <a
            className={clsx('button', styles.buttonSecondary)}
            href="https://github.com/carlosdorneles-mb/mb-cli"
            target="_blank"
            rel="noopener noreferrer">
            GitHub
          </a>
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="MB CLI: documentação oficial. Orquestre plugins com cache SQLite, comandos dinâmicos e ambiente controlado.">
      <HomepageHeader />
      <main>
        <InstallSection />
        <QuickStartSection />
        <KeyFeaturesSection />
        <HomepageFeatures />
        <CallToActionSection />
      </main>
    </Layout>
  );
}

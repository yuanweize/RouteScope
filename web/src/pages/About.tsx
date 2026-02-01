import React, { useEffect, useState } from 'react';
import { Card, Typography, Row, Col, Tag, Button, Space, Divider, Steps, Spin } from 'antd';
import {
  DeploymentUnitOutlined,
  GithubOutlined,
  BugOutlined,
  BookOutlined,
  DashboardOutlined,
  AimOutlined,
  FileTextOutlined,
  HeartFilled,
  CloudDownloadOutlined,
  AppleOutlined,
  WindowsOutlined,
  LinuxOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useTheme } from '../context/ThemeContext';
import { getSystemInfo, type SystemInfo } from '../api';

const { Title, Text, Paragraph } = Typography;

const GITHUB_URL = 'https://github.com/yuanweize/RouteLens';
const RELEASES_URL = `${GITHUB_URL}/releases`;

const About: React.FC = () => {
  const { t } = useTranslation();
  const { isDark } = useTheme();
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchInfo = async () => {
      try {
        const info = await getSystemInfo();
        setSystemInfo(info);
      } catch (e) {
        console.error('Failed to fetch system info:', e);
      } finally {
        setLoading(false);
      }
    };
    fetchInfo();
  }, []);

  // Dynamic version and download URL
  const currentVersion = systemInfo?.version || 'latest';
  // Use GitHub's "latest" redirect for downloads - always gets the newest release
  const latestDownloadBase = `${GITHUB_URL}/releases/latest/download`;

  const techStack = [
    { name: 'Go 1.24', color: '#00ADD8' },
    { name: 'React 19', color: '#61DAFB' },
    { name: 'TypeScript', color: '#3178C6' },
    { name: 'Ant Design v5', color: '#1677ff' },
    { name: 'ECharts', color: '#E43961' },
    { name: 'Vite', color: '#646CFF' },
  ];

  const downloads = [
    { os: 'macOS', arch: 'Apple Silicon', icon: <AppleOutlined />, file: 'routelens-darwin-arm64', color: '#000' },
    { os: 'macOS', arch: 'Intel', icon: <AppleOutlined />, file: 'routelens-darwin-amd64', color: '#555' },
    { os: 'Linux', arch: 'x64', icon: <LinuxOutlined />, file: 'routelens-linux-amd64', color: '#FCC624' },
    { os: 'Linux', arch: 'ARM64', icon: <LinuxOutlined />, file: 'routelens-linux-arm64', color: '#E95420' },
    { os: 'Windows', arch: 'x64', icon: <WindowsOutlined />, file: 'routelens-windows-amd64.exe', color: '#0078D4' },
  ];

  const cardStyle: React.CSSProperties = {
    borderRadius: 12,
    boxShadow: isDark
      ? '0 4px 12px rgba(0, 0, 0, 0.3)'
      : '0 4px 12px rgba(0, 0, 0, 0.08)',
    transition: 'all 0.3s ease',
  };

  const guideSteps = [
    {
      icon: <DashboardOutlined />,
      title: t('about.guide.dashboard'),
      description: t('about.guide.dashboardDesc'),
    },
    {
      icon: <AimOutlined />,
      title: t('about.guide.targets'),
      description: t('about.guide.targetsDesc'),
    },
    {
      icon: <FileTextOutlined />,
      title: t('about.guide.logs'),
      description: t('about.guide.logsDesc'),
    },
  ];

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto', padding: '0 16px' }}>
      {/* Hero Section */}
      <Card style={{ ...cardStyle, marginBottom: 24, overflow: 'hidden' }} bodyStyle={{ padding: 0 }}>
        <div
          style={{
            background: isDark
              ? 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)'
              : 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            padding: '48px 32px',
            textAlign: 'center',
          }}
        >
          <DeploymentUnitOutlined
            style={{
              fontSize: 72,
              color: '#fff',
              marginBottom: 16,
              filter: 'drop-shadow(0 4px 8px rgba(0,0,0,0.2))',
            }}
          />
          <Title level={1} style={{ color: '#fff', marginBottom: 8, fontSize: 36 }}>
            RouteLens
          </Title>
          <Space size="middle" style={{ marginBottom: 16 }}>
            <Tag color="blue" style={{ fontSize: 14, padding: '4px 12px' }}>
              {loading ? <Spin size="small" /> : `v${currentVersion.replace(/^v/, '')}`}
            </Tag>
            <Tag color="green" style={{ fontSize: 14, padding: '4px 12px' }}>
              {t('about.license')}
            </Tag>
          </Space>
          <Paragraph
            style={{
              color: 'rgba(255, 255, 255, 0.9)',
              fontSize: 18,
              maxWidth: 600,
              margin: '0 auto',
            }}
          >
            {t('about.slogan')}
          </Paragraph>
        </div>
      </Card>

      <Row gutter={[24, 24]}>
        {/* Tech Stack */}
        <Col xs={24} md={12}>
          <Card
            title={
              <Space>
                <span style={{ fontSize: 18 }}>🛠️</span>
                <span>{t('about.techStack')}</span>
              </Space>
            }
            style={{ ...cardStyle, height: '100%' }}
            styles={{ body: { height: 'calc(100% - 57px)', display: 'flex', alignItems: 'center' } }}
            hoverable
          >
            <Space wrap size={[8, 12]} style={{ width: '100%' }}>
              {techStack.map((tech) => (
                <Tag
                  key={tech.name}
                  style={{
                    fontSize: 13,
                    padding: '6px 14px',
                    borderRadius: 16,
                    backgroundColor: isDark ? `${tech.color}22` : `${tech.color}15`,
                    color: tech.color,
                    border: `1px solid ${tech.color}40`,
                    fontWeight: 500,
                  }}
                >
                  {tech.name}
                </Tag>
              ))}
            </Space>
          </Card>
        </Col>

        {/* Resource Links */}
        <Col xs={24} md={12}>
          <Card
            title={
              <Space>
                <span style={{ fontSize: 18 }}>🔗</span>
                <span>{t('about.resources')}</span>
              </Space>
            }
            style={{ ...cardStyle, height: '100%' }}
            hoverable
          >
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Button
                type="default"
                icon={<GithubOutlined />}
                href={GITHUB_URL}
                target="_blank"
                block
                size="large"
                style={{ textAlign: 'left', justifyContent: 'flex-start' }}
              >
                {t('about.repository')}
              </Button>
              <Button
                type="default"
                icon={<BugOutlined />}
                href={`${GITHUB_URL}/issues`}
                target="_blank"
                block
                size="large"
                style={{ textAlign: 'left', justifyContent: 'flex-start' }}
              >
                {t('about.reportIssue')}
              </Button>
              <Button
                type="default"
                icon={<BookOutlined />}
                href={`${GITHUB_URL}#readme`}
                target="_blank"
                block
                size="large"
                style={{ textAlign: 'left', justifyContent: 'flex-start' }}
              >
                {t('about.documentation')}
              </Button>
            </Space>
          </Card>
        </Col>

        {/* Downloads / Releases */}
        <Col xs={24}>
          <Card
            title={
              <Space>
                <CloudDownloadOutlined style={{ fontSize: 18, color: '#1677ff' }} />
                <span>{t('about.downloads')}</span>
                <Tag color="blue">{loading ? '...' : `v${currentVersion.replace(/^v/, '')}`}</Tag>
              </Space>
            }
            extra={
              <Button
                type="link"
                href={RELEASES_URL}
                target="_blank"
                icon={<GithubOutlined />}
              >
                {t('about.allReleases')}
              </Button>
            }
            style={cardStyle}
            hoverable
          >
            <Row gutter={[12, 12]}>
              {downloads.map((dl) => (
                <Col xs={24} sm={12} md={8} lg={4} key={dl.file}>
                  <Button
                    type="default"
                    href={`${latestDownloadBase}/${dl.file}`}
                    target="_blank"
                    block
                    style={{
                      height: 'auto',
                      padding: '12px 8px',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: 4,
                      borderRadius: 8,
                      transition: 'all 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.borderColor = dl.color;
                      e.currentTarget.style.color = dl.color;
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.borderColor = '';
                      e.currentTarget.style.color = '';
                    }}
                  >
                    <span style={{ fontSize: 24 }}>{dl.icon}</span>
                    <Text strong style={{ fontSize: 13 }}>{dl.os}</Text>
                    <Text type="secondary" style={{ fontSize: 11 }}>{dl.arch}</Text>
                  </Button>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>

        {/* Quick Guide */}
        <Col xs={24}>
          <Card
            title={
              <Space>
                <span style={{ fontSize: 18 }}>📖</span>
                <span>{t('about.quickGuide')}</span>
              </Space>
            }
            style={cardStyle}
            hoverable
          >
            <Steps
              direction="vertical"
              size="small"
              current={-1}
              items={guideSteps.map((step) => ({
                title: <Text strong style={{ fontSize: 15 }}>{step.title}</Text>,
                description: <Text type="secondary">{step.description}</Text>,
                icon: <span style={{ color: '#1677ff' }}>{step.icon}</span>,
              }))}
            />
          </Card>
        </Col>

        {/* Features */}
        <Col xs={24}>
          <Card
            title={
              <Space>
                <span style={{ fontSize: 18 }}>✨</span>
                <span>{t('about.features')}</span>
              </Space>
            }
            style={cardStyle}
            hoverable
          >
            <Row gutter={[16, 16]}>
              {(t('about.featureList', { returnObjects: true }) as string[]).map((feature, idx) => (
                <Col xs={24} sm={12} key={idx}>
                  <div
                    style={{
                      padding: '12px 16px',
                      borderRadius: 8,
                      backgroundColor: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.02)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                    }}
                  >
                    <span style={{ color: '#52c41a', fontSize: 16 }}>✓</span>
                    <Text>{feature}</Text>
                  </div>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>

      {/* Footer */}
      <Divider style={{ margin: '32px 0 24px' }} />
      <div style={{ textAlign: 'center', paddingBottom: 24 }}>
        <Text type="secondary">
          {t('about.madeWith')} <HeartFilled style={{ color: '#ff4d4f' }} /> {t('about.byAuthor')}
        </Text>
      </div>
    </div>
  );
};

export default About;

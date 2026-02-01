import React from 'react';
import { Card, Typography, Row, Col, Tag, Button, Space, Divider, Steps } from 'antd';
import {
  DeploymentUnitOutlined,
  GithubOutlined,
  BugOutlined,
  BookOutlined,
  DashboardOutlined,
  AimOutlined,
  FileTextOutlined,
  HeartFilled,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useTheme } from '../context/ThemeContext';

const { Title, Text, Paragraph } = Typography;

const GITHUB_URL = 'https://github.com/yuanweize/RouteLens';

const About: React.FC = () => {
  const { t } = useTranslation();
  const { isDark } = useTheme();

  const techStack = [
    { name: 'Go 1.24', color: '#00ADD8' },
    { name: 'React 19', color: '#61DAFB' },
    { name: 'TypeScript', color: '#3178C6' },
    { name: 'Ant Design v5', color: '#1677ff' },
    { name: 'ECharts', color: '#E43961' },
    { name: 'Vite', color: '#646CFF' },
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
              v1.1.0
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
            style={cardStyle}
            hoverable
          >
            <Space wrap size={[8, 12]}>
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
            style={cardStyle}
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

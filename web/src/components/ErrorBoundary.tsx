import { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';
import { Result, Button } from '@arco-design/web-react';

interface Props {
    children: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('Uncaught error:', error, errorInfo);
    }

    public render() {
        if (this.state.hasError) {
            return (
                <div style={{ padding: '40px' }}>
                    <Result
                        status='error'
                        title='Something went wrong'
                        subTitle={this.state.error?.message || 'Unknown Error'}
                        extra={
                            <Button type='primary' onClick={() => window.location.reload()}>
                                Reload Page
                            </Button>
                        }
                    />
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;

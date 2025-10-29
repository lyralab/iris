import React, { useState, useEffect } from 'react';
import config from '../config';
import { useAuth } from '../contexts/AuthContext';
import Layout from '../components/Layout';
import FiringIssues from '../FiringIssues';
import ResolvedIssues from '../ResolvedIssues';

const AlertsPage = () => {
    const [alerts, setAlerts] = useState({ critical: 0, high: 0, medium: 0, low: 0, page: 0, warning: 0 });
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { token } = useAuth();

    useEffect(() => {
        const fetchAlertData = async () => {
            setLoading(true);
            setError(null);

            try {
                const response = await fetch(config.api.alertSummary, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                });

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const data = await response.json();

                const mappedData = {};
                data.severites.forEach((item) => {
                    mappedData[item.severity] = item.count;
                });

                setAlerts(mappedData);
            } catch (e) {
                setError(e.message);
            } finally {
                setLoading(false);
            }
        };

        fetchAlertData();
    }, [token]);

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading Alert Summary...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error">Error fetching Alert Summary: {error}</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="container">
                <div className="alert-summary">
                    <h2>Alert Summary</h2>
                    <div className="summary-items">
                        <div className="summary-item critical">
                            <div className="icon">🔥</div>
                            <div className="count">{alerts.critical || 0}</div>
                            <div className="label">Critical</div>
                        </div>
                        <div className="summary-item high">
                            <div className="icon">⚠️</div>
                            <div className="count">{alerts.high || 0}</div>
                            <div className="label">High</div>
                        </div>
                        <div className="summary-item medium">
                            <div className="icon">🟡</div>
                            <div className="count">{alerts.medium || 0}</div>
                            <div className="label">Medium</div>
                        </div>
                        <div className="summary-item warning">
                            <div className="icon">🟡</div>
                            <div className="count">{alerts.warning || 0}</div>
                            <div className="label">Warning</div>
                        </div>
                        <div className="summary-item low">
                            <div className="icon">🟢</div>
                            <div className="count">{alerts.low || 0}</div>
                            <div className="label">Low</div>
                        </div>
                        <div className="summary-item page">
                            <div className="icon">✉️</div>
                            <div className="count">{alerts.page || 0}</div>
                            <div className="label">Page</div>
                        </div>
                    </div>
                </div>
                <FiringIssues />
                <ResolvedIssues />
            </div>
        </Layout>
    );
};

export default AlertsPage;

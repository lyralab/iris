import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import './AlertsPage.css';

const AlertsPage = () => {
    const [alerts, setAlerts] = useState({ critical: 0, high: 0, medium: 0, low: 0, page: 0, warning: 0 });
    const [allAlerts, setAllAlerts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('summary');

    useEffect(() => {
        fetchAllAlertData();
    }, []);

    const fetchAllAlertData = async () => {
        setLoading(true);
        setError(null);

        try {
            // Fetch alert summary
            const summaryData = await apiService.getAlertSummary();
            const mappedData = {};

            // Handle null or undefined severites array
            if (summaryData && summaryData.severites && Array.isArray(summaryData.severites)) {
                summaryData.severites.forEach(item => {
                    if (item && item.severity && typeof item.count === 'number') {
                        mappedData[item.severity] = item.count;
                    }
                });
            }
            setAlerts(mappedData);

            // Fetch all alerts - handle null response
            const alertsData = await apiService.getAlerts({ limit: 100, page: 1 });
            setAllAlerts(alertsData && alertsData.alerts && Array.isArray(alertsData.alerts) ? alertsData.alerts : []);

        } catch (e) {
            setError(e.message);
            // Set empty arrays on error to prevent crashes
            setAllAlerts([]);
        } finally {
            setLoading(false);
        }
    };

    const fetchFilteredAlerts = async (status, severity = '') => {
        try {
            const data = await apiService.getAlerts({ status, severity, limit: 100, page: 1 });
            // Handle null or missing alerts array
            setAllAlerts(data && data.alerts && Array.isArray(data.alerts) ? data.alerts : []);
        } catch (e) {
            console.error('Error fetching filtered alerts:', e);
            // Set empty array on error
            setAllAlerts([]);
        }
    };

    const getSeverityClass = (severity) => {
        const classes = {
            critical: 'severity-critical',
            high: 'severity-high',
            medium: 'severity-medium',
            warning: 'severity-warning',
            low: 'severity-low',
            page: 'severity-page',
        };
        return classes[severity?.toLowerCase()] || 'severity-default';
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading alerts...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading alerts: {error}</div>
            </Layout>
        );
    }

    const firingAlerts = Array.isArray(allAlerts) ? allAlerts.filter(a => a && a.status === 'firing') : [];
    const resolvedAlerts = Array.isArray(allAlerts) ? allAlerts.filter(a => a && a.status === 'resolved') : [];

    return (
        <Layout>
            <div className="alerts-page">
                <h1>Alerts Dashboard</h1>

                <div className="tabs">
                    <button
                        className={`tab ${activeTab === 'summary' ? 'active' : ''}`}
                        onClick={() => setActiveTab('summary')}
                    >
                        Summary
                    </button>
                    <button
                        className={`tab ${activeTab === 'firing' ? 'active' : ''}`}
                        onClick={() => {
                            setActiveTab('firing');
                            fetchFilteredAlerts('firing');
                        }}
                    >
                        Firing Alerts ({firingAlerts.length})
                    </button>
                    <button
                        className={`tab ${activeTab === 'resolved' ? 'active' : ''}`}
                        onClick={() => {
                            setActiveTab('resolved');
                            fetchFilteredAlerts('resolved');
                        }}
                    >
                        Resolved Alerts ({resolvedAlerts.length})
                    </button>
                </div>

                {activeTab === 'summary' && (
                    <div className="alert-summary">
                        <h2>Alert Summary by Severity</h2>
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
                )}

                {activeTab === 'firing' && (
                    <div className="alerts-list">
                        <h2>Firing Alerts</h2>
                        {firingAlerts.length === 0 ? (
                            <p className="no-alerts">No firing alerts</p>
                        ) : (
                            <div className="alerts-table-container">
                                <table className="alerts-table">
                                    <thead>
                                        <tr>
                                            <th>Alert Name</th>
                                            <th>Severity</th>
                                            <th>Status</th>
                                            <th>Started At</th>
                                            <th>Description</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {firingAlerts.map((alert, index) => (
                                            <tr key={index}>
                                                <td className="alert-name">{alert.alert_name || 'Unknown'}</td>
                                                <td>
                                                    <span className={`severity-badge ${getSeverityClass(alert.severity)}`}>
                                                        {alert.severity || 'N/A'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="status-badge firing">Firing</span>
                                                </td>
                                                <td>{new Date(alert.starts_at).toLocaleString()}</td>
                                                <td className="alert-description">{alert.summary || 'No description'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'resolved' && (
                    <div className="alerts-list">
                        <h2>Resolved Alerts</h2>
                        {resolvedAlerts.length === 0 ? (
                            <p className="no-alerts">No resolved alerts</p>
                        ) : (
                            <div className="alerts-table-container">
                                <table className="alerts-table">
                                    <thead>
                                        <tr>
                                            <th>Alert Name</th>
                                            <th>Severity</th>
                                            <th>Status</th>
                                            <th>Resolved At</th>
                                            <th>Description</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {resolvedAlerts.map((alert, index) => (
                                            <tr key={index}>
                                                <td className="alert-name">{alert.alert_name || 'Unknown'}</td>
                                                <td>
                                                    <span className={`severity-badge ${getSeverityClass(alert.severity)}`}>
                                                        {alert.severity || 'N/A'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="status-badge resolved">Resolved</span>
                                                </td>
                                                <td>{new Date(alert.ends_at).toLocaleString()}</td>
                                                <td className="alert-description">{alert.summary || 'No description'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </Layout>
    );
};

export default AlertsPage;


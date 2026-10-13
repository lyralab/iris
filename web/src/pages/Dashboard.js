import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import useAutoRefresh, { REFRESH_OPTIONS, readRefreshMs, formatLastUpdated } from '../utils/useAutoRefresh';
import './Dashboard.css';

function timeValue(value) {
    const t = new Date(value).getTime();
    return Number.isNaN(t) ? 0 : t;
}

const SEVERITIES = [
    { key: 'critical', label: 'Critical' },
    { key: 'high', label: 'High' },
    { key: 'warning', label: 'Warning' },
    { key: 'medium', label: 'Medium' },
    { key: 'low', label: 'Low' },
    { key: 'page', label: 'Page' },
];

const Dashboard = () => {
    const [alerts, setAlerts] = useState({ critical: 0, high: 0, medium: 0, low: 0, page: 0, warning: 0 });
    const [topAlerts, setTopAlerts] = useState([]);
    const [resolvedAlerts, setResolvedAlerts] = useState([]);
    const [firingLimit, setFiringLimit] = useState(10);
    const [resolvedLimit, setResolvedLimit] = useState(10);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [refreshMs, setRefreshMs] = useState(() => readRefreshMs('iris.dashboard.refreshMs'));
    const [lastUpdated, setLastUpdated] = useState(null);
    const [refreshError, setRefreshError] = useState(null);

    const fetchDashboardData = useCallback(async (silent = false) => {
        if (!silent) {
            setLoading(true);
            setError(null);
        } else {
            setRefreshError(null);
        }

        try {
            // Fetch alert summary
            const alertData = await apiService.getAlertSummary();
            const mappedData = {};

            // Handle null or undefined severites array
            if (alertData && alertData.severites && Array.isArray(alertData.severites)) {
                alertData.severites.forEach(item => {
                    if (item && item.severity && typeof item.count === 'number') {
                        mappedData[item.severity] = item.count;
                    }
                });
            }

            setAlerts(mappedData);

            // Fetch detailed firing alerts for critical and warning
            try {
                const detailedAlerts = await apiService.getAlerts({ status: 'firing', limit: firingLimit * 2, page: 1 });
                if (detailedAlerts && detailedAlerts.alerts && Array.isArray(detailedAlerts.alerts)) {
                    const filteredAlerts = detailedAlerts.alerts
                        .filter(alert => alert && (alert.severity === 'critical' || alert.severity === 'warning'))
                        .sort((a, b) => timeValue(b.starts_at || b.created_at) - timeValue(a.starts_at || a.created_at))
                        .slice(0, firingLimit);
                    setTopAlerts(filteredAlerts);
                }
            } catch (e) {
                console.log('Detailed alerts API error:', e);
                setTopAlerts([]);
            }

            // Fetch resolved alerts
            try {
                const resolvedData = await apiService.getAlerts({ status: 'resolved', limit: resolvedLimit, page: 1 });
                if (resolvedData && resolvedData.alerts && Array.isArray(resolvedData.alerts)) {
                    setResolvedAlerts(
                        [...resolvedData.alerts].sort(
                            (a, b) => timeValue(b.ends_at || b.updated_at) - timeValue(a.ends_at || a.updated_at),
                        ),
                    );
                }
            } catch (e) {
                console.log('Resolved alerts API error:', e);
                setResolvedAlerts([]);
            }

            setLastUpdated(new Date());
        } catch (e) {
            if (silent) {
                setRefreshError(e.message);
            } else {
                setError(e.message);
            }
        } finally {
            if (!silent) {
                setLoading(false);
            }
        }
    }, [firingLimit, resolvedLimit]);

    useEffect(() => {
        fetchDashboardData(false);
    }, [fetchDashboardData]);

    useAutoRefresh(() => fetchDashboardData(true), refreshMs);

    const onRefreshChange = (value) => {
        setRefreshMs(value);
        try {
            localStorage.setItem('iris.dashboard.refreshMs', String(value));
        } catch (e) {
            // ignore persistence errors
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
                <div className="loading">Loading Dashboard...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading dashboard: {error}</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="dashboard">
                <div className="page-header">
                    <h1>Dashboard</h1>
                    <div className="header-actions refresh-controls">
                        {lastUpdated && (
                            <span className="last-updated">Last updated {formatLastUpdated(lastUpdated)}</span>
                        )}
                        <label className="filter-field">
                            Refresh
                            <select
                                className="select-input"
                                value={refreshMs}
                                onChange={(e) => onRefreshChange(Number(e.target.value))}
                            >
                                {REFRESH_OPTIONS.map((o) => (
                                    <option key={o.value} value={o.value}>{o.label}</option>
                                ))}
                            </select>
                        </label>
                        <button className="btn-secondary" onClick={() => fetchDashboardData(true)}>
                            Refresh
                        </button>
                    </div>
                </div>
                {refreshError && <div className="refresh-hint">Background refresh failed: {refreshError}</div>}

                <section className="active-alerts">
                    <h2>Active alerts</h2>
                    <div className="stats-grid">
                        {SEVERITIES.map((item) => (
                            <Link
                                key={item.key}
                                to={`/alerts?severity=${item.key}`}
                                className={`stat-card stat-${item.key}`}
                            >
                                <span className="stat-label">{item.label}</span>
                                <span className="stat-value">{alerts[item.key] || 0}</span>
                            </Link>
                        ))}
                    </div>
                </section>

                <div className="top-alerts-section">
                        <div className="section-header">
                            <h2>Latest Critical & Warning Alerts ({topAlerts.length})</h2>
                            <div className="alert-config">
                                <label htmlFor="firing-limit">Show:</label>
                                <select
                                    id="firing-limit"
                                    value={firingLimit}
                                    onChange={(e) => setFiringLimit(Number(e.target.value))}
                                    className="alert-limit-dropdown"
                                >
                                    <option value={5}>5</option>
                                    <option value={10}>10</option>
                                    <option value={15}>15</option>
                                    <option value={20}>20</option>
                                </select>
                            </div>
                        </div>
                        {topAlerts.length === 0 ? (
                            <p className="empty">No firing alerts</p>
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
                                        {topAlerts.map((alert, index) => (
                                            <tr key={alert.id || index}>
                                                <td className="alert-name">{alert.name || alert.alert_name || 'Unknown'}</td>
                                                <td>
                                                    <span className={`severity-badge ${getSeverityClass(alert.severity)}`}>
                                                        {alert.severity || 'N/A'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="status-badge firing">Firing</span>
                                                </td>
                                                <td>{new Date(alert.starts_at || alert.created_at).toLocaleString()}</td>
                                                <td className="alert-description">{alert.description || alert.summary || 'No description'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                </div>

                <div className="resolved-alerts-section">
                        <div className="section-header">
                            <h2>Latest Resolved Alerts ({resolvedAlerts.length})</h2>
                            <div className="alert-config">
                                <label htmlFor="resolved-limit">Show:</label>
                                <select
                                    id="resolved-limit"
                                    value={resolvedLimit}
                                    onChange={(e) => setResolvedLimit(Number(e.target.value))}
                                    className="alert-limit-dropdown"
                                >
                                    <option value={5}>5</option>
                                    <option value={10}>10</option>
                                    <option value={15}>15</option>
                                    <option value={20}>20</option>
                                </select>
                            </div>
                        </div>
                        {resolvedAlerts.length === 0 ? (
                            <p className="empty">No resolved alerts</p>
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
                                            <tr key={alert.id || index}>
                                                <td className="alert-name">{alert.name || alert.alert_name || 'Unknown'}</td>
                                                <td>
                                                    <span className={`severity-badge ${getSeverityClass(alert.severity)}`}>
                                                        {alert.severity || 'N/A'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="status-badge resolved">Resolved</span>
                                                </td>
                                                <td>{new Date(alert.ends_at || alert.updated_at).toLocaleString()}</td>
                                                <td className="alert-description">{alert.description || alert.summary || 'No description'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                </div>
            </div>
        </Layout>
    );
};

export default Dashboard;


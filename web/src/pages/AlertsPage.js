import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import './AlertsPage.css';

const SEVERITIES = ['critical', 'high', 'warning', 'medium', 'low', 'page'];

const AlertsPage = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const [allAlerts, setAllAlerts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('firing');
    const severity = searchParams.get('severity') || '';

    useEffect(() => {
        let cancelled = false;

        const load = async () => {
            setLoading(true);
            setError(null);
            try {
                const data = await apiService.getAlerts({
                    status: activeTab,
                    severity,
                    limit: 100,
                    page: 1,
                });
                const list = data && data.alerts && Array.isArray(data.alerts) ? data.alerts : [];
                if (!cancelled) {
                    setAllAlerts(list);
                }
            } catch (e) {
                if (!cancelled) {
                    setError(e.message);
                    setAllAlerts([]);
                }
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        };

        load();
        return () => {
            cancelled = true;
        };
    }, [activeTab, severity]);

    const onSeverityChange = (value) => {
        const next = new URLSearchParams(searchParams);
        if (value) {
            next.set('severity', value);
        } else {
            next.delete('severity');
        }
        setSearchParams(next);
    };

    const getSeverityClass = (level) => {
        const classes = {
            critical: 'severity-critical',
            high: 'severity-high',
            medium: 'severity-medium',
            warning: 'severity-warning',
            low: 'severity-low',
            page: 'severity-page',
        };
        return classes[level?.toLowerCase()] || 'severity-default';
    };

    const rows = Array.isArray(allAlerts) ? allAlerts.filter(alert => alert && alert.status === activeTab) : [];
    const timeLabel = activeTab === 'firing' ? 'Started At' : 'Resolved At';
    const timeField = activeTab === 'firing' ? 'starts_at' : 'ends_at';

    return (
        <Layout>
            <div className="alerts-page">
                <div className="page-header">
                    <h1>Alerts</h1>
                    <label className="filter-field">
                        Severity
                        <select
                            className="select-input"
                            value={severity}
                            onChange={(e) => onSeverityChange(e.target.value)}
                        >
                            <option value="">All</option>
                            {SEVERITIES.map((level) => (
                                <option key={level} value={level}>
                                    {level.charAt(0).toUpperCase() + level.slice(1)}
                                </option>
                            ))}
                        </select>
                    </label>
                </div>

                <div className="tabs">
                    <button
                        className={`tab ${activeTab === 'firing' ? 'active' : ''}`}
                        onClick={() => setActiveTab('firing')}
                    >
                        Firing
                    </button>
                    <button
                        className={`tab ${activeTab === 'resolved' ? 'active' : ''}`}
                        onClick={() => setActiveTab('resolved')}
                    >
                        Resolved
                    </button>
                </div>

                {error && <div className="error-message">Error loading alerts: {error}</div>}

                <div className="alerts-list">
                    {loading ? (
                        <div className="loading">Loading alerts...</div>
                    ) : rows.length === 0 ? (
                        <p className="empty">No {activeTab} alerts</p>
                    ) : (
                        <div className="alerts-table-container">
                            <table className="alerts-table">
                                <thead>
                                    <tr>
                                        <th>Alert Name</th>
                                        <th>Severity</th>
                                        <th>Status</th>
                                        <th>{timeLabel}</th>
                                        <th>Description</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {rows.map((alert, index) => (
                                        <tr key={alert.id || index}>
                                            <td className="alert-name">{alert.alert_name || alert.name || 'Unknown'}</td>
                                            <td>
                                                <span className={`severity-badge ${getSeverityClass(alert.severity)}`}>
                                                    {alert.severity || 'N/A'}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`status-badge ${activeTab}`}>{activeTab}</span>
                                            </td>
                                            <td>{alert[timeField] ? new Date(alert[timeField]).toLocaleString() : '-'}</td>
                                            <td className="alert-description">{alert.summary || alert.description || 'No description'}</td>
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

export default AlertsPage;

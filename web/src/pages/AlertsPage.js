import React, { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import useAutoRefresh, { REFRESH_OPTIONS, readRefreshMs, formatLastUpdated } from '../utils/useAutoRefresh';
import './AlertsPage.css';

const SEVERITIES = ['critical', 'high', 'warning', 'medium', 'low', 'page'];

const AlertsPage = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const [allAlerts, setAllAlerts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('firing');
    const severity = searchParams.get('severity') || '';
    const [refreshMs, setRefreshMs] = useState(() => readRefreshMs('iris.alerts.refreshMs'));
    const [lastUpdated, setLastUpdated] = useState(null);
    const [refreshError, setRefreshError] = useState(null);

    const load = useCallback(async (silent = false) => {
        if (!silent) {
            setLoading(true);
            setError(null);
        } else {
            setRefreshError(null);
        }
        try {
            const data = await apiService.getAlerts({
                status: activeTab,
                severity,
                limit: 100,
                page: 1,
            });
            const list = data && data.alerts && Array.isArray(data.alerts) ? data.alerts : [];
            setAllAlerts(list);
            setLastUpdated(new Date());
        } catch (e) {
            if (silent) {
                setRefreshError(e.message);
            } else {
                setError(e.message);
                setAllAlerts([]);
            }
        } finally {
            if (!silent) {
                setLoading(false);
            }
        }
    }, [activeTab, severity]);

    useEffect(() => {
        load(false);
    }, [load]);

    useAutoRefresh(() => load(true), refreshMs);

    const onRefreshChange = (value) => {
        setRefreshMs(value);
        try {
            localStorage.setItem('iris.alerts.refreshMs', String(value));
        } catch (e) {
            // ignore persistence errors
        }
    };

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
                    <div className="header-actions refresh-controls">
                        {lastUpdated && (
                            <span className="last-updated">Last updated {formatLastUpdated(lastUpdated)}</span>
                        )}
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
                        <button className="btn-secondary" onClick={() => load(true)}>
                            Refresh
                        </button>
                    </div>
                </div>
                {refreshError && <div className="refresh-hint">Background refresh failed: {refreshError}</div>}

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

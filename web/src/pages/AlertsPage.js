import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import AlertList from '../components/AlertList';
import EmptyState from '../components/EmptyState';
import Icon from '../components/Icon';
import RefreshControls from '../components/RefreshControls';
import useAutoRefresh, { readRefreshMs, latestOnly } from '../utils/useAutoRefresh';
import { byNewest } from '../utils/alertSort';
import { SEVERITIES } from '../utils/format';
import './AlertsPage.css';

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
    const [query, setQuery] = useState('');
    const [startRequest] = useState(latestOnly);
    const manualPending = useRef(false);

    const load = useCallback(async (silent = false) => {
        // A background refresh must not supersede a pending manual load and hide its error.
        if (silent && manualPending.current) return;
        const isLatest = startRequest();
        if (!silent) {
            manualPending.current = true;
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
            if (!isLatest()) return;
            const list = data && data.alerts && Array.isArray(data.alerts) ? data.alerts : [];
            setAllAlerts(list);
            setLastUpdated(new Date());
        } catch (e) {
            if (!isLatest()) return;
            if (silent) {
                setRefreshError(e.message);
            } else {
                setError(e.message);
                setAllAlerts([]);
            }
        } finally {
            if (isLatest()) {
                manualPending.current = false;
                setLoading(false);
            }
        }
    }, [activeTab, severity, startRequest]);

    useEffect(() => {
        load(false);
    }, [load]);

    const refresh = useAutoRefresh(() => load(true), refreshMs);

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

    const needle = query.trim().toLowerCase();
    const rows = (Array.isArray(allAlerts) ? allAlerts.filter(alert => alert && alert.status === activeTab) : [])
        .filter((alert) => !needle || [alert.name, alert.alert_name, alert.description, alert.summary]
            .join(' ').toLowerCase().includes(needle))
        .sort(byNewest);
    const severityLabel = SEVERITIES.find((s) => s.key === severity)?.label;

    let empty;
    if (needle) {
        empty = <EmptyState icon="search" title="No matching alerts">Nothing matches “{query.trim()}”.</EmptyState>;
    } else if (activeTab === 'firing') {
        empty = (
            <EmptyState icon="checkCircle" tone="success" title={`No ${severityLabel ? severityLabel.toLowerCase() + ' ' : ''}alerts firing`}>
                Everything is quiet right now.
            </EmptyState>
        );
    } else {
        empty = <EmptyState icon="inbox" title={`No resolved ${severityLabel ? severityLabel.toLowerCase() + ' ' : ''}alerts`} />;
    }

    return (
        <Layout>
            <div className="page alerts-page">
                <div className="page-header">
                    <div>
                        <h1>Alerts</h1>
                        <p className="subtitle">Everything received from Alertmanager and Grafana</p>
                    </div>
                    <RefreshControls
                        refreshMs={refreshMs}
                        onRefreshMsChange={onRefreshChange}
                        onRefresh={refresh}
                        lastUpdated={lastUpdated}
                    />
                </div>

                {refreshError && (
                    <div className="refresh-hint"><Icon name="alert" />Background refresh failed: {refreshError}</div>
                )}

                <section className="card alerts-card">
                    <div className="alerts-toolbar">
                        <div className="segmented" role="tablist" aria-label="Alert status">
                            <button
                                type="button"
                                role="tab"
                                aria-selected={activeTab === 'firing'}
                                className={activeTab === 'firing' ? 'active' : ''}
                                onClick={() => setActiveTab('firing')}
                            >
                                <Icon name="flame" size={14} />
                                Firing
                            </button>
                            <button
                                type="button"
                                role="tab"
                                aria-selected={activeTab === 'resolved'}
                                className={activeTab === 'resolved' ? 'active' : ''}
                                onClick={() => setActiveTab('resolved')}
                            >
                                <Icon name="checkCircle" size={14} />
                                Resolved
                            </button>
                        </div>
                        <div className="input-with-icon alerts-search">
                            <Icon name="search" />
                            <input
                                className="search-input"
                                type="search"
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                placeholder="Search alerts"
                                aria-label="Search alerts"
                            />
                        </div>
                    </div>

                    <div className="alerts-filters">
                        <div className="chips" role="group" aria-label="Filter by severity">
                            <button
                                type="button"
                                className={`chip ${severity === '' ? 'active' : ''}`}
                                aria-pressed={severity === ''}
                                onClick={() => onSeverityChange('')}
                            >
                                All severities
                            </button>
                            {SEVERITIES.map((level) => (
                                <button
                                    type="button"
                                    key={level.key}
                                    className={`chip sev-${level.key} ${severity === level.key ? 'active' : ''}`}
                                    aria-pressed={severity === level.key}
                                    onClick={() => onSeverityChange(severity === level.key ? '' : level.key)}
                                >
                                    <span className="sev-dot" />
                                    {level.label}
                                </button>
                            ))}
                        </div>
                        {!loading && !error && (
                            <span className="alerts-count tabular">
                                {rows.length} {rows.length === 1 ? 'alert' : 'alerts'}
                            </span>
                        )}
                    </div>

                    {error && (
                        <div className="alerts-error">
                            <div className="error-message"><Icon name="alert" />Error loading alerts: {error}</div>
                        </div>
                    )}

                    {!error && <AlertList alerts={rows} loading={loading} empty={empty} />}
                </section>
            </div>
        </Layout>
    );
};

export default AlertsPage;

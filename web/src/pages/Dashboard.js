import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import AlertList from '../components/AlertList';
import EmptyState from '../components/EmptyState';
import Icon from '../components/Icon';
import RefreshControls from '../components/RefreshControls';
import useAutoRefresh, { readRefreshMs, latestOnly } from '../utils/useAutoRefresh';
import { byNewest } from '../utils/alertSort';
import { SEVERITIES } from '../utils/format';
import './Dashboard.css';

const LIMITS = [5, 10, 15, 20];

const LimitSelect = ({ id, value, onChange }) => (
    <label className="field-inline" htmlFor={id}>
        Show
        <select id={id} className="select-input select-sm" value={value} onChange={(e) => onChange(Number(e.target.value))}>
            {LIMITS.map((n) => <option key={n} value={n}>{n}</option>)}
        </select>
    </label>
);

const SeverityBar = ({ counts, total }) => (
    <div className="severity-bar" role="img" aria-label={SEVERITIES.map((s) => `${s.label} ${counts[s.key] || 0}`).join(', ')}>
        {total === 0 ? (
            <span className="severity-bar-empty" />
        ) : (
            SEVERITIES.filter((s) => counts[s.key] > 0).map((s) => (
                <span
                    key={s.key}
                    className={`severity-bar-seg sev-${s.key}`}
                    style={{ flexGrow: counts[s.key] }}
                    title={`${s.label}: ${counts[s.key]}`}
                />
            ))
        )}
    </div>
);

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
    const [startRequest] = useState(latestOnly);
    const manualPending = useRef(false);

    const fetchDashboardData = useCallback(async (silent = false) => {
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
            const alertData = await apiService.getAlertSummary();
            const detailedAlerts = await apiService.getAlerts({ status: 'firing', limit: firingLimit * 2, page: 1 })
                .catch((e) => {
                    console.log('Detailed alerts API error:', e);
                    return null;
                });
            const resolvedData = await apiService.getAlerts({ status: 'resolved', limit: resolvedLimit, page: 1 })
                .catch((e) => {
                    console.log('Resolved alerts API error:', e);
                    return null;
                });
            if (!isLatest()) return;

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

            // Critical and warning firing alerts only
            const firing = detailedAlerts && Array.isArray(detailedAlerts.alerts) ? detailedAlerts.alerts : [];
            setTopAlerts(
                firing
                    .filter(alert => alert && (alert.severity === 'critical' || alert.severity === 'warning'))
                    .sort(byNewest)
                    .slice(0, firingLimit),
            );

            const resolved = resolvedData && Array.isArray(resolvedData.alerts) ? resolvedData.alerts : [];
            setResolvedAlerts([...resolved].sort(byNewest));

            setLastUpdated(new Date());
        } catch (e) {
            if (!isLatest()) return;
            if (silent) {
                setRefreshError(e.message);
            } else {
                setError(e.message);
            }
        } finally {
            if (isLatest()) {
                manualPending.current = false;
                setLoading(false);
            }
        }
    }, [firingLimit, resolvedLimit, startRequest]);

    useEffect(() => {
        fetchDashboardData(false);
    }, [fetchDashboardData]);

    const refresh = useAutoRefresh(() => fetchDashboardData(true), refreshMs);

    const onRefreshChange = (value) => {
        setRefreshMs(value);
        try {
            localStorage.setItem('iris.dashboard.refreshMs', String(value));
        } catch (e) {
            // ignore persistence errors
        }
    };

    const total = SEVERITIES.reduce((sum, s) => sum + (alerts[s.key] || 0), 0);
    const urgent = (alerts.critical || 0) + (alerts.high || 0);
    const health = total === 0 ? 'ok' : urgent > 0 ? 'critical' : 'warning';

    return (
        <Layout>
            <div className="page dashboard">
                <div className="page-header">
                    <div>
                        <h1>Overview</h1>
                        <p className="subtitle">Live state of firing alerts across all sources</p>
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
                {error && (
                    <div className="error-message"><Icon name="alert" />Error loading dashboard: {error}</div>
                )}

                <section className="overview-grid" aria-label="Firing alerts by severity">
                    <div className={`card health-card health-${loading ? 'loading' : health}`}>
                        <div className="health-top">
                            <span className="health-icon">
                                <Icon name={health === 'ok' ? 'checkCircle' : 'flame'} size={18} />
                            </span>
                            <span className="health-label">
                                {loading ? 'Checking…' : health === 'ok' ? 'All systems quiet' : `${urgent} critical or high`}
                            </span>
                        </div>
                        <div className="health-total">
                            {loading ? <span className="skeleton" style={{ width: 70, height: 44 }} /> : <span className="tabular">{total}</span>}
                            <span className="health-total-label">firing {total === 1 ? 'alert' : 'alerts'}</span>
                        </div>
                        <SeverityBar counts={alerts} total={loading ? 0 : total} />
                        <Link to="/alerts" className="health-link">
                            View all alerts <Icon name="chevronRight" size={14} />
                        </Link>
                    </div>

                    <div className="stats-grid">
                        {SEVERITIES.map((item) => {
                            const count = alerts[item.key] || 0;
                            return (
                                <Link
                                    key={item.key}
                                    to={`/alerts?severity=${item.key}`}
                                    className={`stat-card sev-${item.key} ${count === 0 ? 'is-zero' : ''}`}
                                >
                                    <span className="stat-label">
                                        <span className="sev-dot" />
                                        {item.label}
                                    </span>
                                    <span className="stat-value tabular">
                                        {loading ? <span className="skeleton" style={{ width: 32, height: 28 }} /> : count}
                                    </span>
                                </Link>
                            );
                        })}
                    </div>
                </section>

                <div className="dashboard-columns">
                    <section className="card">
                        <div className="card-header">
                            <h2 className="card-title">
                                <Icon name="flame" />
                                Needs attention
                                <span className="count">{loading ? '' : topAlerts.length}</span>
                            </h2>
                            <LimitSelect id="firing-limit" value={firingLimit} onChange={setFiringLimit} />
                        </div>
                        <AlertList
                            alerts={topAlerts}
                            loading={loading}
                            empty={
                                <EmptyState icon="checkCircle" tone="success" title="Nothing critical is firing">
                                    Critical and warning alerts will show up here.
                                </EmptyState>
                            }
                        />
                    </section>

                    <section className="card">
                        <div className="card-header">
                            <h2 className="card-title">
                                <Icon name="checkCircle" />
                                Recently resolved
                                <span className="count">{loading ? '' : resolvedAlerts.length}</span>
                            </h2>
                            <LimitSelect id="resolved-limit" value={resolvedLimit} onChange={setResolvedLimit} />
                        </div>
                        <AlertList
                            alerts={resolvedAlerts}
                            loading={loading}
                            empty={<EmptyState icon="inbox" title="No resolved alerts yet" />}
                        />
                    </section>
                </div>
            </div>
        </Layout>
    );
};

export default Dashboard;

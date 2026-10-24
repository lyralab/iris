import React, { useEffect, useState } from 'react';
import Icon from './Icon';
import EmptyState from './EmptyState';
import { severityKey, severityClass, timeAgo, formatDuration, formatDateTime } from '../utils/format';
import { timeOrNull } from '../utils/alertSort';
import { splitDescription } from '../utils/alertDescription';
import './AlertList.css';

// Re-renders every 30s so relative times ("5m ago") stay current between data refreshes.
function useNow(intervalMs = 30000) {
    const [now, setNow] = useState(Date.now);
    useEffect(() => {
        const id = setInterval(() => setNow(Date.now()), intervalMs);
        return () => clearInterval(id);
    }, [intervalMs]);
    return now;
}

// Labels already shown elsewhere in the row, or used only for Iris routing.
const HIDDEN_LABELS = new Set(['alertname', 'severity', 'method', 'receptor']);
// Annotations rendered in dedicated places instead of the generic list.
const SPECIAL_ANNOTATIONS = new Set(['summary', 'description']);

// Only link to http(s) URLs so a crafted payload can't inject javascript: links.
const safeUrl = (url) => (/^https?:\/\/\S+$/i.test(url || '') ? url : null);

// "runbook_url" -> "Runbook", "grafana_dashboard" -> "Grafana dashboard".
const linkLabel = (key) => {
    const words = key.replace(/[_-]?url$/i, '').replace(/[_-]+/g, ' ').trim() || 'Link';
    return words.charAt(0).toUpperCase() + words.slice(1);
};

const AlertRow = ({ alert, now, expanded, onToggle }) => {
    const resolved = alert.status === 'resolved';
    const sev = severityKey(alert.severity);
    const name = alert.name || alert.alert_name || 'Unknown alert';
    const description = alert.description || alert.summary || '';
    const started = alert.starts_at || alert.created_at;
    const ended = alert.ends_at || alert.updated_at;
    const when = resolved ? ended : started;
    const duration = resolved ? formatDuration(started, ended) : formatDuration(started, now);
    const hasEnd = timeOrNull(alert.ends_at) !== null;
    const annotations = alert.annotations || {};
    const { text: fullDescription, value: alertValue } = splitDescription(annotations.description || description);
    const labels = Object.entries(alert.labels || {})
        .filter(([key]) => !HIDDEN_LABELS.has(key.toLowerCase()))
        .sort(([a], [b]) => a.localeCompare(b));
    const otherAnnotations = Object.entries(annotations)
        .filter(([key, value]) => value && !SPECIAL_ANNOTATIONS.has(key))
        .sort(([a], [b]) => a.localeCompare(b));
    // Annotations holding a URL (runbook, Grafana dashboard, ...) become links; the rest stay as fields.
    const extraAnnotations = otherAnnotations.filter(([, value]) => !safeUrl(value));
    const links = otherAnnotations
        .filter(([, value]) => safeUrl(value))
        .sort(([a], [b]) => (b === 'runbook_url') - (a === 'runbook_url'))
        .map(([key, value]) => ({ key, href: value, label: linkLabel(key) }));
    const sourceUrl = safeUrl(alert.generator_url);
    if (sourceUrl) links.push({ key: 'generator_url', href: sourceUrl, label: 'View in Prometheus' });

    return (
        <li className={`alert-row sev-${sev} ${resolved ? 'is-resolved' : 'is-firing'} ${expanded ? 'is-expanded' : ''}`}>
            <button type="button" className="alert-row-main" onClick={onToggle} aria-expanded={expanded}>
                <span className={`sev-dot ${!resolved && (sev === 'critical' || sev === 'high') ? 'pulse' : ''}`} />
                <span className="alert-row-text">
                    <span className="alert-row-name">{name}</span>
                    <span className="alert-row-desc" title={description}>{description || 'No description'}</span>
                </span>
                <span className="alert-row-meta">
                    <span className={`severity-badge ${severityClass(alert.severity)}`}>{alert.severity || 'unknown'}</span>
                    <span className="alert-row-time" title={formatDateTime(when)}>
                        {resolved ? <Icon name="checkCircle" size={13} /> : <Icon name="clock" size={13} />}
                        {timeAgo(when, now) || '—'}
                    </span>
                </span>
                <Icon name="chevronDown" className="alert-row-chevron" />
            </button>
            {expanded && (
                <div className="alert-row-details">
                    {fullDescription && <p className="alert-row-fulldesc">{fullDescription}</p>}
                    <dl>
                        <div><dt>Status</dt><dd><span className={`status-badge ${resolved ? 'resolved' : 'firing'}`}>{resolved ? 'Resolved' : 'Firing'}</span></dd></div>
                        <div><dt>Started</dt><dd>{formatDateTime(started)}</dd></div>
                        <div><dt>{resolved ? 'Resolved' : 'Ends'}</dt><dd>{hasEnd || resolved ? formatDateTime(ended) : '—'}</dd></div>
                        <div><dt>{resolved ? 'Lasted' : 'Firing for'}</dt><dd>{duration || '—'}</dd></div>
                        {alertValue && <div><dt>Value</dt><dd className="mono">{alertValue}</dd></div>}
                        {alert.id && <div><dt>ID</dt><dd className="mono">{alert.id}</dd></div>}
                        {extraAnnotations.map(([key, value]) => (
                            <div key={key}><dt>{key.replace(/_/g, ' ')}</dt><dd>{value}</dd></div>
                        ))}
                    </dl>
                    {labels.length > 0 && (
                        <div className="alert-row-labels">
                            <span className="alert-row-section">Labels</span>
                            <ul>
                                {labels.map(([key, value]) => (
                                    <li key={key} className="label-chip mono"><span>{key}</span>={value}</li>
                                ))}
                            </ul>
                        </div>
                    )}
                    {links.length > 0 && (
                        <div className="alert-row-links">
                            {links.map(({ key, href, label }) => (
                                <a key={key} href={href} target="_blank" rel="noopener noreferrer">{label}</a>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </li>
    );
};

const SkeletonRows = ({ count = 5 }) => (
    <div aria-hidden="true">
        {Array.from({ length: count }, (_, i) => (
            <div className="skeleton-row" key={i}>
                <span className="skeleton" style={{ width: 8, height: 8, borderRadius: 4 }} />
                <span style={{ flex: 1, display: 'grid', gap: 6 }}>
                    <span className="skeleton" style={{ width: `${40 + ((i * 17) % 35)}%`, height: 12 }} />
                    <span className="skeleton" style={{ width: `${60 + ((i * 11) % 30)}%`, height: 10 }} />
                </span>
                <span className="skeleton" style={{ width: 64, height: 20, borderRadius: 999 }} />
            </div>
        ))}
    </div>
);

const AlertList = ({ alerts, loading = false, empty }) => {
    const now = useNow();
    const [open, setOpen] = useState(() => new Set());

    if (loading) return <SkeletonRows />;
    if (!alerts.length) return empty || <EmptyState icon="inbox" title="No alerts" />;

    const toggle = (key) => {
        setOpen((prev) => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key);
            else next.add(key);
            return next;
        });
    };

    return (
        <ul className="alert-list">
            {alerts.map((alert, index) => {
                const key = alert.id || `${alert.name}-${index}`;
                return (
                    <AlertRow
                        key={key}
                        alert={alert}
                        now={now}
                        expanded={open.has(key)}
                        onToggle={() => toggle(key)}
                    />
                );
            })}
        </ul>
    );
};

export default AlertList;

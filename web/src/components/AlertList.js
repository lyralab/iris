import React, { useEffect, useState } from 'react';
import Icon from './Icon';
import EmptyState from './EmptyState';
import { severityKey, severityClass, timeAgo, formatDuration, formatDateTime } from '../utils/format';
import { timeOrNull } from '../utils/alertSort';
import { splitDescription } from '../utils/alertDescription';
import { isAdmin } from '../utils/auth';
import apiService from '../utils/apiService';
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

// Durations offered to admins; '' means until an admin unsilences (server caps timed ones at 30 days).
export const SILENCE_DURATIONS = [
    { value: '1h', label: '1 hour' },
    { value: '4h', label: '4 hours' },
    { value: '24h', label: '24 hours' },
    { value: '168h', label: '7 days' },
    { value: '', label: 'Until unsilenced' },
];

// The server already drops expired silences; re-check so one that expires between refreshes clears.
export const isSilenced = (alert, now = Date.now()) => {
    if (!alert.silenced) return false;
    const until = Date.parse(alert.silenced_until || '');
    return Number.isNaN(until) || until > now;
};

const silenceText = (alert) => {
    const by = alert.silenced_by ? ` by ${alert.silenced_by}` : '';
    const until = alert.silenced_until ? `until ${formatDateTime(alert.silenced_until)}` : 'until unsilenced';
    return `Notifications off for everyone${by}, ${until}`;
};

// Admin-only. A silence stops notifications for every user, not just the admin.
const SilenceControls = ({ alert, silenced, onSilenceChange }) => {
    const [duration, setDuration] = useState('1h');
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');

    const run = async (request) => {
        setBusy(true);
        setError('');
        try {
            const res = await request();
            if (res && res.alert) onSilenceChange(res.alert);
        } catch (e) {
            setError(e.message || 'Request failed');
        } finally {
            setBusy(false);
        }
    };

    return (
        <div className="alert-row-silence">
            {silenced ? (
                <button type="button" className="btn-secondary" disabled={busy}
                    onClick={() => run(() => apiService.unsilenceAlert(alert.id))}>
                    <Icon name="bell" size={14} /> Unsilence
                </button>
            ) : (
                <>
                    <select aria-label="Silence duration" value={duration} disabled={busy}
                        onChange={(e) => setDuration(e.target.value)}>
                        {SILENCE_DURATIONS.map((d) => <option key={d.value} value={d.value}>{d.label}</option>)}
                    </select>
                    <button type="button" className="btn-secondary" disabled={busy}
                        onClick={() => run(() => apiService.silenceAlert(alert.id, duration))}>
                        <Icon name="bellOff" size={14} /> Silence for everyone
                    </button>
                </>
            )}
            {error && <span className="alert-row-silence-error" role="alert">{error}</span>}
        </div>
    );
};

const AlertRow = ({ alert, now, expanded, onToggle, admin, onSilenceChange }) => {
    const resolved = alert.status === 'resolved';
    const silenced = isSilenced(alert, now);
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
        <li className={`alert-row sev-${sev} ${resolved ? 'is-resolved' : 'is-firing'} ${silenced ? 'is-silenced' : ''} ${expanded ? 'is-expanded' : ''}`}>
            <button type="button" className="alert-row-main" onClick={onToggle} aria-expanded={expanded}>
                <span className={`sev-dot ${!resolved && (sev === 'critical' || sev === 'high') ? 'pulse' : ''}`} />
                <span className="alert-row-text">
                    <span className="alert-row-name">{name}</span>
                    <span className="alert-row-desc" title={description}>{description || 'No description'}</span>
                </span>
                <span className="alert-row-meta">
                    {silenced && (
                        <span className="silenced-badge" title={silenceText(alert)}>
                            <Icon name="bellOff" size={12} /> Silenced
                        </span>
                    )}
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
                        {silenced && <div><dt>Silenced</dt><dd>{silenceText(alert)}</dd></div>}
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
                    {admin && alert.id && (!resolved || silenced) && (
                        <SilenceControls alert={alert} silenced={silenced} onSilenceChange={onSilenceChange} />
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

// onSilenceChange (optional) runs after an admin silences or unsilences, e.g. to reload a filtered list.
const AlertList = ({ alerts, loading = false, empty, onSilenceChange }) => {
    const now = useNow();
    const [open, setOpen] = useState(() => new Set());
    // Silence changes made here, shown until the next data refresh brings the server's view.
    const [silenceOverrides, setSilenceOverrides] = useState({});
    useEffect(() => setSilenceOverrides({}), [alerts]);
    const admin = isAdmin();

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
                        alert={silenceOverrides[alert.id] ? { ...alert, ...silenceOverrides[alert.id] } : alert}
                        admin={admin}
                        onSilenceChange={(updated) => {
                            setSilenceOverrides((prev) => ({
                                ...prev,
                                [alert.id]: {
                                    silenced: updated.silenced,
                                    silenced_until: updated.silenced_until,
                                    silenced_by: updated.silenced_by,
                                },
                            }));
                            if (onSilenceChange) onSilenceChange(updated);
                        }}
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

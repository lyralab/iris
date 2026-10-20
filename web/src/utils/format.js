import { timeOrNull } from './alertSort';

export const SEVERITIES = [
    { key: 'critical', label: 'Critical' },
    { key: 'high', label: 'High' },
    { key: 'warning', label: 'Warning' },
    { key: 'medium', label: 'Medium' },
    { key: 'low', label: 'Low' },
    { key: 'page', label: 'Page' },
];

const KNOWN = new Set(SEVERITIES.map((s) => s.key));

export function severityKey(severity) {
    const key = String(severity || '').toLowerCase();
    return KNOWN.has(key) ? key : 'default';
}

export function severityClass(severity) {
    return `severity-${severityKey(severity)}`;
}

const USER_STATUS = {
    verified: { label: 'Active', className: 'status-active' },
    new: { label: 'New', className: 'status-new' },
    disable: { label: 'Disabled', className: 'status-disable' },
    disabled: { label: 'Disabled', className: 'status-disable' },
};

export function userStatusInfo(status) {
    return USER_STATUS[String(status || '').toLowerCase()] || { label: status || 'Unknown', className: 'status-unknown' };
}

// "just now", "5m ago", "3h ago", "2d ago"; '' for missing or Go zero times.
export function timeAgo(value, now = Date.now()) {
    const t = timeOrNull(value);
    if (t === null) return '';
    const s = Math.max(0, Math.round((now - t) / 1000));
    if (s < 45) return 'just now';
    const m = Math.round(s / 60);
    if (m < 60) return `${m}m ago`;
    const h = Math.round(m / 60);
    if (h < 24) return `${h}h ago`;
    const d = Math.round(h / 24);
    if (d < 30) return `${d}d ago`;
    return new Date(t).toLocaleDateString();
}

// Compact span between two times: "45s", "12m", "3h 20m", "4d 6h"; '' if either is missing.
export function formatDuration(from, to = Date.now()) {
    const a = timeOrNull(from);
    const b = typeof to === 'number' ? to : timeOrNull(to);
    if (a === null || b === null) return '';
    const s = Math.max(0, Math.round((b - a) / 1000));
    if (s < 60) return `${s}s`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m`;
    const h = Math.floor(m / 60);
    if (h < 24) return m % 60 ? `${h}h ${m % 60}m` : `${h}h`;
    const d = Math.floor(h / 24);
    return h % 24 ? `${d}d ${h % 24}h` : `${d}d`;
}

export function formatDateTime(value) {
    const t = timeOrNull(value);
    return t === null ? '—' : new Date(t).toLocaleString();
}

export function initials(...parts) {
    const words = parts.filter(Boolean).join(' ').trim().split(/\s+/).filter(Boolean);
    if (words.length === 0) return '?';
    if (words.length === 1) return words[0].slice(0, 2);
    return words[0][0] + words[1][0];
}

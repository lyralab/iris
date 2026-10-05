import { useEffect, useRef } from 'react';

export const REFRESH_OPTIONS = [
    { value: 0, label: 'Off' },
    { value: 15000, label: '15s' },
    { value: 30000, label: '30s' },
    { value: 60000, label: '1m' },
    { value: 300000, label: '5m' },
];

export const DEFAULT_REFRESH_MS = 30000;

export function readRefreshMs(key) {
    try {
        const raw = localStorage.getItem(key);
        if (raw === null) return DEFAULT_REFRESH_MS;
        const ms = Number(raw);
        return REFRESH_OPTIONS.some((o) => o.value === ms) ? ms : DEFAULT_REFRESH_MS;
    } catch (e) {
        return DEFAULT_REFRESH_MS;
    }
}

// ponytail: fixed-interval polling ceiling; upgrade path is backoff/jitter if polling load becomes an issue.
export default function useAutoRefresh(callback, intervalMs) {
    const saved = useRef(callback);
    useEffect(() => {
        saved.current = callback;
    });

    useEffect(() => {
        if (!intervalMs || intervalMs <= 0) return undefined;
        const id = setInterval(() => {
            if (document.hidden) return;
            saved.current();
        }, intervalMs);
        return () => clearInterval(id);
    }, [intervalMs]);
}

export function formatLastUpdated(date) {
    if (!date) return '';
    try {
        return new Date(date).toLocaleTimeString();
    } catch (e) {
        return '';
    }
}

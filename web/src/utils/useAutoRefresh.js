import { useEffect, useRef, useState } from 'react';

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

// Wraps an async fn so calls made while one is still pending are dropped.
export function skipWhileRunning(fn) {
    let running = false;
    return async (...args) => {
        if (running) return;
        running = true;
        try {
            await fn(...args);
        } finally {
            running = false;
        }
    };
}

// Each call starts a request and returns isLatest(), false once a newer request has started.
export function latestOnly() {
    let seq = 0;
    return () => {
        const id = ++seq;
        return () => id === seq;
    };
}

// Returns the guarded refresh so manual Refresh buttons share the one-in-flight slot with the timer.
// ponytail: fixed-interval polling ceiling; upgrade path is backoff/jitter if polling load becomes an issue.
export default function useAutoRefresh(callback, intervalMs) {
    const saved = useRef(callback);
    useEffect(() => {
        saved.current = callback;
    });
    const [tick] = useState(() => skipWhileRunning(() => saved.current()));

    useEffect(() => {
        if (!intervalMs || intervalMs <= 0) return undefined;
        const id = setInterval(() => {
            if (document.hidden) return;
            tick();
        }, intervalMs);
        return () => clearInterval(id);
    }, [intervalMs, tick]);

    return tick;
}

export function formatLastUpdated(date) {
    if (!date) return '';
    try {
        return new Date(date).toLocaleTimeString();
    } catch (e) {
        return '';
    }
}

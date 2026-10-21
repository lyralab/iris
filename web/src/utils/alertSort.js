// The API sends Go's zero time ("0001-01-01 00:00:00 +0000 UTC") for a NULL column.
// Match it by prefix: V8 parses that string as year 2001, not year 1.
export function timeOrNull(value) {
    if (!value || String(value).startsWith('0001-01-01')) return null;
    const t = new Date(value).getTime();
    return Number.isNaN(t) ? null : t;
}

function firstTime(alert, fields) {
    for (const field of fields) {
        const t = timeOrNull(alert[field]);
        if (t !== null) return t;
    }
    return 0;
}

// Mirrors the ORDER BY in pkg/storage/postgresql/alert_repository.go.
export function alertSortKey(alert) {
    return alert.status === 'resolved'
        ? firstTime(alert, ['ends_at', 'updated_at', 'created_at'])
        : firstTime(alert, ['starts_at', 'created_at']);
}

export function byNewest(a, b) {
    return alertSortKey(b) - alertSortKey(a) || firstTime(b, ['created_at']) - firstTime(a, ['created_at']);
}

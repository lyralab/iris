import { alertSortKey, byNewest } from './alertSort';

// What the API sends for a NULL/zero column (Go time.Time.String()).
const ZERO = '0001-01-01 00:00:00 +0000 UTC';

test('ISO zero time also counts as missing', () => {
    expect(alertSortKey({ status: 'firing', starts_at: '0001-01-01T00:00:00Z', created_at: '2026-01-01T00:00:00Z' }))
        .toBe(Date.parse('2026-01-01T00:00:00Z'));
});

test('parses the Go time.Time.String() format the API sends', () => {
    expect(alertSortKey({ status: 'firing', starts_at: '2026-01-02 10:00:00 +0000 UTC' }))
        .toBe(Date.parse('2026-01-02T10:00:00Z'));
});

test('firing alerts fall back from starts_at to created_at', () => {
    expect(alertSortKey({ status: 'firing', starts_at: '2026-01-02T00:00:00Z', created_at: '2026-01-01T00:00:00Z' }))
        .toBe(Date.parse('2026-01-02T00:00:00Z'));
    expect(alertSortKey({ status: 'firing', starts_at: ZERO, created_at: '2026-01-01T00:00:00Z' }))
        .toBe(Date.parse('2026-01-01T00:00:00Z'));
    expect(alertSortKey({ status: 'firing', created_at: '2026-01-01T00:00:00Z' }))
        .toBe(Date.parse('2026-01-01T00:00:00Z'));
});

test('resolved alerts fall back from ends_at to updated_at to created_at', () => {
    const created = '2026-01-01T00:00:00Z';
    const updated = '2026-01-03T00:00:00Z';
    expect(alertSortKey({ status: 'resolved', ends_at: '2026-01-05T00:00:00Z', updated_at: updated, created_at: created }))
        .toBe(Date.parse('2026-01-05T00:00:00Z'));
    expect(alertSortKey({ status: 'resolved', ends_at: ZERO, updated_at: updated, created_at: created }))
        .toBe(Date.parse(updated));
    expect(alertSortKey({ status: 'resolved', ends_at: 'garbage', created_at: created }))
        .toBe(Date.parse(created));
});

test('byNewest orders like the server, tie-broken by created_at', () => {
    const rows = [
        { id: 'old', status: 'firing', starts_at: '2026-01-01T00:00:00Z' },
        { id: 'no-start', status: 'firing', starts_at: ZERO, created_at: '2026-01-04T00:00:00Z' },
        { id: 'tie-older', status: 'firing', starts_at: '2026-01-02T00:00:00Z', created_at: '2026-01-01T00:00:00Z' },
        { id: 'tie-newer', status: 'firing', starts_at: '2026-01-02T00:00:00Z', created_at: '2026-01-02T00:00:00Z' },
    ];
    expect([...rows].sort(byNewest).map((r) => r.id)).toEqual(['no-start', 'tie-newer', 'tie-older', 'old']);
});

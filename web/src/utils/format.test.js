import { timeAgo, formatDuration, severityClass, userStatusInfo, initials } from './format';

const NOW = Date.parse('2026-01-01T12:00:00Z');

test('timeAgo buckets by unit and ignores Go zero time', () => {
    expect(timeAgo('2026-01-01T11:59:50Z', NOW)).toBe('just now');
    expect(timeAgo('2026-01-01T11:55:00Z', NOW)).toBe('5m ago');
    expect(timeAgo('2026-01-01T09:00:00Z', NOW)).toBe('3h ago');
    expect(timeAgo('2025-12-30T12:00:00Z', NOW)).toBe('2d ago');
    expect(timeAgo('0001-01-01 00:00:00 +0000 UTC', NOW)).toBe('');
    expect(timeAgo(null, NOW)).toBe('');
});

test('formatDuration renders the two largest units', () => {
    expect(formatDuration('2026-01-01T11:59:15Z', NOW)).toBe('45s');
    expect(formatDuration('2026-01-01T11:48:00Z', NOW)).toBe('12m');
    expect(formatDuration('2026-01-01T08:40:00Z', NOW)).toBe('3h 20m');
    expect(formatDuration('2026-01-01T10:00:00Z', NOW)).toBe('2h');
    expect(formatDuration('2025-12-28T06:00:00Z', NOW)).toBe('4d 6h');
    expect(formatDuration('2026-01-01T10:00:00Z', '0001-01-01 00:00:00 +0000 UTC')).toBe('');
});

test('severityClass falls back for unknown levels', () => {
    expect(severityClass('Critical')).toBe('severity-critical');
    expect(severityClass('info')).toBe('severity-default');
    expect(severityClass(undefined)).toBe('severity-default');
});

test('userStatusInfo is case-insensitive', () => {
    expect(userStatusInfo('Verified').label).toBe('Active');
    expect(userStatusInfo('new').className).toBe('status-new');
    expect(userStatusInfo('weird')).toEqual({ label: 'weird', className: 'status-unknown' });
});

test('initials', () => {
    expect(initials('Ada', 'Lovelace')).toBe('AL');
    expect(initials('admin')).toBe('ad');
    expect(initials('', null)).toBe('?');
});

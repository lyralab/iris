import { splitDescription } from './alertDescription';

test('drops the Go map LABELS line and lifts VALUE into its own field', () => {
    const text = 'Disk is almost full (< 10% left)\n  VALUE = 6.2\n  LABELS = map[device:/dev/sda1 mountpoint:/var/lib/docker]';
    expect(splitDescription(text)).toEqual({ text: 'Disk is almost full (< 10% left)', value: '6.2' });
});

test('leaves ordinary descriptions untouched', () => {
    const text = 'Job batch/nightly-report is taking more than 12 hours to complete.';
    expect(splitDescription(text)).toEqual({ text, value: null });
});

test('keeps a LABELS line that is not a Go map', () => {
    expect(splitDescription('LABELS = see dashboard').text).toBe('LABELS = see dashboard');
});

test('handles missing descriptions', () => {
    expect(splitDescription(undefined)).toEqual({ text: '', value: null });
});

// Many Prometheus rules (e.g. awesome-prometheus-alerts) end their description with
//   VALUE = {{ $value }}
//   LABELS = {{ $labels }}
// which renders as a raw Go map ("map[device:/dev/sda1 mountpoint:/var]").
// The labels are already shown as chips, so drop that line and lift VALUE into its own field.
const LABELS_LINE = /^\s*LABELS\s*=\s*map\[.*\]\s*$/i;
const VALUE_LINE = /^\s*VALUE\s*=\s*(.*?)\s*$/i;

export function splitDescription(text) {
    let value = null;
    const lines = [];
    for (const line of String(text || '').split('\n')) {
        if (LABELS_LINE.test(line)) continue;
        const match = value === null && line.match(VALUE_LINE);
        if (match && match[1]) {
            value = match[1];
            continue;
        }
        lines.push(line.trim());
    }
    return { text: lines.join('\n').trim(), value };
}

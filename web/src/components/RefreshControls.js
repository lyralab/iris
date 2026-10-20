import React, { useState } from 'react';
import Icon from './Icon';
import { REFRESH_OPTIONS, formatLastUpdated } from '../utils/useAutoRefresh';

// Live indicator + interval picker + manual refresh, shared by the alert views.
const RefreshControls = ({ refreshMs, onRefreshMsChange, onRefresh, lastUpdated }) => {
    const [spinning, setSpinning] = useState(false);
    const live = refreshMs > 0;

    const click = async () => {
        setSpinning(true);
        try {
            await onRefresh();
        } finally {
            setSpinning(false);
        }
    };

    return (
        <div className="refresh-controls">
            <span className={`live-indicator ${live ? '' : 'paused'}`} title={lastUpdated ? `Last updated ${formatLastUpdated(lastUpdated)}` : undefined}>
                <span className={`sev-dot ${live ? 'pulse' : ''}`} />
                {live ? 'Live' : 'Paused'}
                {lastUpdated && <span className="last-updated">· {formatLastUpdated(lastUpdated)}</span>}
            </span>
            <select
                className="select-input select-sm"
                value={refreshMs}
                onChange={(e) => onRefreshMsChange(Number(e.target.value))}
                aria-label="Auto-refresh interval"
            >
                {REFRESH_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>{o.value ? `Every ${o.label}` : 'Auto-refresh off'}</option>
                ))}
            </select>
            <button type="button" className="btn-secondary btn-sm" onClick={click} disabled={spinning}>
                <Icon name="refresh" size={14} className={spinning ? 'spin' : ''} />
                Refresh
            </button>
        </div>
    );
};

export default RefreshControls;

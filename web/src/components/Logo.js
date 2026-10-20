import React from 'react';

// The Iris "signal arcs" mark: two arcs broadcasting from a source dot.
// Arcs use currentColor so the mark follows the surrounding text color; the dot takes the accent.
// `tile` draws it on a rounded square, matching the favicon and app icons in public/.
const Logo = ({ size = 28, tile = false, dotColor, title, className = '' }) => (
    <svg
        className={`logo-mark ${className}`.trim()}
        width={size}
        height={size}
        viewBox="0 0 64 64"
        role={title ? 'img' : undefined}
        aria-label={title}
        aria-hidden={title ? undefined : 'true'}
        focusable="false"
    >
        {tile && <rect width="64" height="64" rx="14" fill="var(--logo-tile, #0f172a)" />}
        <g
            fill="none"
            stroke={tile ? 'var(--logo-tile-ink, #ffffff)' : 'currentColor'}
            strokeLinecap="round"
            strokeWidth={tile ? 5.5 : 6.5}
        >
            {tile ? (
                <>
                    <path d="M12 41.5A20 20 0 0 1 52 41.5" />
                    <path d="M21 41.5A11 11 0 0 1 43 41.5" />
                </>
            ) : (
                <>
                    <path d="M5 44A27 27 0 0 1 59 44" />
                    <path d="M17 44A15 15 0 0 1 47 44" />
                </>
            )}
        </g>
        <circle cx="32" cy={tile ? 41.5 : 44} r={tile ? 4.5 : 5.5} fill={dotColor || (tile ? 'var(--logo-dot, #818cf8)' : 'var(--accent)')} />
    </svg>
);

export default Logo;

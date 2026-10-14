import React, { useState } from 'react';
import { getTheme, setTheme } from '../utils/theme';

const ThemeToggle = () => {
    const [theme, setCurrent] = useState(getTheme);
    const next = theme === 'dark' ? 'light' : 'dark';

    const toggle = () => {
        setTheme(next);
        setCurrent(next);
    };

    return (
        <button
            type="button"
            className="theme-toggle"
            onClick={toggle}
            aria-label={`Switch to ${next} theme`}
            title={`Switch to ${next} theme`}
        >
            {theme === 'dark' ? '☀️' : '🌙'}
        </button>
    );
};

export default ThemeToggle;

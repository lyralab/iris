import React, { useEffect, useState } from 'react';
import { getTheme, setTheme } from '../utils/theme';
import Icon from './Icon';

const ThemeToggle = () => {
    const [theme, setCurrent] = useState(getTheme);
    const next = theme === 'dark' ? 'light' : 'dark';

    // Keeps every toggle on the page (sidebar + mobile top bar) showing the same icon.
    useEffect(() => {
        const sync = () => setCurrent(getTheme());
        window.addEventListener('iris:theme', sync);
        return () => window.removeEventListener('iris:theme', sync);
    }, []);

    const toggle = () => {
        setTheme(next);
        setCurrent(next);
        window.dispatchEvent(new Event('iris:theme'));
    };

    return (
        <button
            type="button"
            className="theme-toggle"
            onClick={toggle}
            aria-label={`Switch to ${next} theme`}
            title={`Switch to ${next} theme`}
        >
            <Icon name={theme === 'dark' ? 'sun' : 'moon'} />
        </button>
    );
};

export default ThemeToggle;

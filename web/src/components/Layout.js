import React, { useCallback, useEffect, useState } from 'react';
import { NavLink, useNavigate, useLocation } from 'react-router-dom';
import { logout, getUserFromToken, isAdmin } from '../utils/auth';
import { initials } from '../utils/format';
import ThemeToggle from './ThemeToggle';
import Icon from './Icon';
import Logo from './Logo';
import './Layout.css';

const MAIN_NAV = [
    { to: '/dashboard', label: 'Overview', icon: 'dashboard' },
    { to: '/alerts', label: 'Alerts', icon: 'bell' },
];

const ADMIN_NAV = [
    { to: '/users', label: 'Users', icon: 'users' },
    { to: '/groups', label: 'Groups', icon: 'group' },
    { to: '/providers', label: 'Providers', icon: 'send' },
];

const COLLAPSED_KEY = 'iris.sidebar.collapsed';

function readCollapsed() {
    try {
        return localStorage.getItem(COLLAPSED_KEY) === '1';
    } catch {
        return false;
    }
}

const NavItem = ({ item }) => (
    <li>
        <NavLink to={item.to} data-label={item.label} aria-label={item.label} className={({ isActive }) => `nav-link ${isActive ? 'active' : ''}`}>
            <Icon name={item.icon} size={17} />
            <span>{item.label}</span>
        </NavLink>
    </li>
);

const Layout = ({ children }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const user = getUserFromToken();
    const userIsAdmin = isAdmin();
    const [drawerOpen, setDrawerOpen] = useState(false);
    const [collapsed, setCollapsed] = useState(readCollapsed);

    useEffect(() => {
        setDrawerOpen(false);
    }, [location.pathname]);

    const toggleCollapsed = useCallback(() => {
        setCollapsed((prev) => {
            const next = !prev;
            try {
                localStorage.setItem(COLLAPSED_KEY, next ? '1' : '0');
            } catch {
                // private mode: the choice lasts for this page only
            }
            return next;
        });
    }, []);

    // "[" toggles the sidebar, unless the user is typing in a field.
    useEffect(() => {
        const onKey = (e) => {
            if (e.key !== '[' || e.metaKey || e.ctrlKey || e.altKey) return;
            const el = e.target;
            if (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName)) return;
            toggleCollapsed();
        };
        document.addEventListener('keydown', onKey);
        return () => document.removeEventListener('keydown', onKey);
    }, [toggleCollapsed]);

    const handleLogout = () => {
        logout();
        navigate('/');
    };

    const username = user?.username || 'User';

    return (
        <div className={`app-shell ${drawerOpen ? 'drawer-open' : ''} ${collapsed ? 'sidebar-collapsed' : ''}`}>
            <aside className="sidebar" aria-label="Main navigation">
                <button
                    type="button"
                    className="sidebar-toggle"
                    onClick={toggleCollapsed}
                    aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                    aria-expanded={!collapsed}
                    title={`${collapsed ? 'Expand' : 'Collapse'} sidebar ([)`}
                >
                    <Icon name={collapsed ? 'chevronRight' : 'chevronLeft'} size={14} strokeWidth={2.2} />
                </button>
                <div className="sidebar-brand">
                    <Logo tile size={30} />
                    <div className="sidebar-brand-text">
                        <span className="brand-name">Iris</span>
                        <span className="brand-sub">Alert Manager</span>
                    </div>
                </div>

                <nav className="sidebar-nav">
                    <span className="nav-section">Monitor</span>
                    <ul>
                        {MAIN_NAV.map((item) => <NavItem key={item.to} item={item} />)}
                    </ul>
                    {userIsAdmin && (
                        <>
                            <span className="nav-section">Administration</span>
                            <ul>
                                {ADMIN_NAV.map((item) => <NavItem key={item.to} item={item} />)}
                            </ul>
                        </>
                    )}
                </nav>

                <div className="sidebar-footer">
                    <NavLink to="/profile" data-label={username} aria-label={`Profile: ${username}`} className={({ isActive }) => `user-card ${isActive ? 'active' : ''}`}>
                        <span className="avatar">{initials(username)}</span>
                        <span className="user-card-text">
                            <span className="user-card-name">{username}</span>
                            <span className="user-card-role">{user?.role || 'member'}</span>
                        </span>
                    </NavLink>
                    <div className="sidebar-actions">
                        <ThemeToggle />
                        <button type="button" className="btn-ghost btn-icon btn-sm" onClick={handleLogout} title="Sign out" aria-label="Sign out">
                            <Icon name="logout" />
                        </button>
                    </div>
                </div>
            </aside>

            <div className="drawer-backdrop" onClick={() => setDrawerOpen(false)} aria-hidden="true" />

            <div className="app-main">
                <header className="topbar">
                    <button
                        type="button"
                        className="btn-ghost btn-icon"
                        onClick={() => setDrawerOpen(true)}
                        aria-label="Open navigation"
                    >
                        <Icon name="menu" size={18} />
                    </button>
                    <div className="topbar-brand">
                        <Logo tile size={26} />
                        <span className="brand-name">Iris</span>
                    </div>
                    <ThemeToggle />
                </header>
                <main className="main-content">
                    {children}
                </main>
            </div>
        </div>
    );
};

export default Layout;

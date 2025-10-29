import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import './Layout.css';

const Layout = ({ children }) => {
    const { user, logout, isAdmin } = useAuth();
    const location = useLocation();

    const isActive = (path) => {
        return location.pathname === path ? 'active' : '';
    };

    return (
        <div className="layout">
            <nav className="navbar">
                <div className="navbar-brand">
                    <h1>Iris Alert System</h1>
                </div>
                <ul className="navbar-menu">
                    {isAdmin() && (
                        <>
                            <li className={isActive('/alerts')}>
                                <Link to="/alerts">Alerts</Link>
                            </li>
                            <li className={isActive('/users')}>
                                <Link to="/users">Users</Link>
                            </li>
                            <li className={isActive('/groups')}>
                                <Link to="/groups">Groups</Link>
                            </li>
                        </>
                    )}
                </ul>
                <div className="navbar-user">
                    <span className="user-info">
                        {user?.username} ({user?.role})
                    </span>
                    <button className="logout-btn" onClick={logout}>
                        Logout
                    </button>
                </div>
            </nav>
            <main className="main-content">{children}</main>
        </div>
    );
};

export default Layout;

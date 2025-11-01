import React from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { logout, getUserFromToken, isAdmin } from '../utils/auth';
import './Layout.css';

const Layout = ({ children }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const user = getUserFromToken();
    const userIsAdmin = isAdmin();

    const handleLogout = () => {
        logout();
        navigate('/');
    };

    const handleProfileClick = () => {
        navigate('/profile');
    };

    return (
        <div className="layout">
            <nav className="navbar">
                <div className="navbar-brand">
                    <img src="/iris.png" alt="Iris Logo" className="logo" />
                    <span className="brand-name">Iris Alert Manager</span>
                </div>
                <ul className="navbar-menu">
                    <li className={location.pathname === '/dashboard' ? 'active' : ''}>
                        <Link to="/dashboard">Dashboard</Link>
                    </li>
                    <li className={location.pathname === '/alerts' ? 'active' : ''}>
                        <Link to="/alerts">Alerts</Link>
                    </li>
                    {userIsAdmin && (
                        <>
                            <li className={location.pathname === '/users' ? 'active' : ''}>
                                <Link to="/users">Users</Link>
                            </li>
                            <li className={location.pathname === '/groups' ? 'active' : ''}>
                                <Link to="/groups">Groups</Link>
                            </li>
                            <li className={location.pathname === '/providers' ? 'active' : ''}>
                                <Link to="/providers">Providers</Link>
                            </li>
                        </>
                    )}
                </ul>
                <div className="navbar-user">
                    <span
                        className="user-name clickable"
                        onClick={handleProfileClick}
                        title="View Profile"
                    >
                        👤 {user?.username || 'User'}
                    </span>
                    <button onClick={handleLogout} className="logout-btn">Logout</button>
                </div>
            </nav>
            <main className="main-content">
                {children}
            </main>
        </div>
    );
};

export default Layout;


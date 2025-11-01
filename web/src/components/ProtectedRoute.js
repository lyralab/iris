import React from 'react';
import { Navigate } from 'react-router-dom';
import { isAuthenticated, isAdmin } from '../utils/auth';

const ProtectedRoute = ({ children, requireAdmin = false }) => {
    if (!isAuthenticated()) {
        return <Navigate to="/" replace />;
    }

    if (requireAdmin && !isAdmin()) {
        return (
            <div style={{ padding: '20px', textAlign: 'center' }}>
                <h2>Access Denied</h2>
                <p>You don't have permission to access this page.</p>
                <p>Admin access required.</p>
            </div>
        );
    }

    return children;
};

export default ProtectedRoute;


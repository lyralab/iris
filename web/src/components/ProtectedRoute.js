import React from 'react';
import { Navigate } from 'react-router-dom';
import { isAuthenticated, isAdmin, mustChangePassword } from '../utils/auth';
import Layout from './Layout';
import EmptyState from './EmptyState';

const ProtectedRoute = ({ children, requireAdmin = false }) => {
    if (!isAuthenticated()) {
        return <Navigate to="/" replace />;
    }

    if (mustChangePassword()) {
        return <Navigate to="/change-password" replace />;
    }

    if (requireAdmin && !isAdmin()) {
        return (
            <Layout>
                <div className="page">
                    <div className="card">
                        <EmptyState icon="lock" title="Admin access required">
                            You don't have permission to view this page.
                        </EmptyState>
                    </div>
                </div>
            </Layout>
        );
    }

    return children;
};

export default ProtectedRoute;


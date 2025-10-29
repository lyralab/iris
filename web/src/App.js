import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import LoginPage from './LoginPage';
import AlertsPage from './pages/AlertsPage';
import UsersPage from './pages/UsersPage';
import GroupsPage from './pages/GroupsPage';
import UnauthorizedPage from './pages/UnauthorizedPage';
import './App.css';

function App() {
    return (
        <Router>
            <AuthProvider>
                <Routes>
                    <Route path="/" element={<LoginPage />} />
                    <Route path="/unauthorized" element={<UnauthorizedPage />} />
                    <Route
                        path="/alerts"
                        element={
                            <ProtectedRoute adminOnly={true}>
                                <AlertsPage />
                            </ProtectedRoute>
                        }
                    />
                    <Route
                        path="/users"
                        element={
                            <ProtectedRoute adminOnly={true}>
                                <UsersPage />
                            </ProtectedRoute>
                        }
                    />
                    <Route
                        path="/groups"
                        element={
                            <ProtectedRoute adminOnly={true}>
                                <GroupsPage />
                            </ProtectedRoute>
                        }
                    />
                    <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
            </AuthProvider>
        </Router>
    );
}

export default App;
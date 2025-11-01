import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import './Profile.css';

const Profile = () => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchUserProfile();
    }, []);

    const fetchUserProfile = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getUserMe();
            setUser(data.user || data);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const getStatusInfo = (status) => {
        const statusMap = {
            'Verified': { label: 'Active', className: 'status-active' },
            'verified': { label: 'Active', className: 'status-active' },
            'New': { label: 'New', className: 'status-new' },
            'new': { label: 'New', className: 'status-new' },
            'Disable': { label: 'Disable', className: 'status-disable' },
            'disable': { label: 'Disable', className: 'status-disable' },
            'disabled': { label: 'Disable', className: 'status-disable' }
        };
        return statusMap[status] || { label: status || 'Unknown', className: 'status-unknown' };
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading profile...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading profile: {error}</div>
            </Layout>
        );
    }

    if (!user) {
        return (
            <Layout>
                <div className="error-message">User information not available</div>
            </Layout>
        );
    }

    const statusInfo = getStatusInfo(user.status);

    return (
        <Layout>
            <div className="profile-page">
                <div className="profile-container">
                    <div className="profile-card">
                        <div className="profile-left">
                            <div className="profile-avatar">
                                <span className="avatar-icon">👤</span>
                            </div>
                            <div className="profile-user-info">
                                <h1>{user.username}</h1>
                                <span className={`status-badge ${statusInfo.className}`}>
                                    {statusInfo.label}
                                </span>
                            </div>
                        </div>

                        <div className="profile-right">
                            <div className="profile-section">
                                <h3>Personal Information</h3>
                                <div className="info-list">
                                    <div className="info-row">
                                        <label>Username</label>
                                        <span>{user.username}</span>
                                    </div>
                                    <div className="info-row">
                                        <label>First Name</label>
                                        <span>{user.firstName || user.firstname || '-'}</span>
                                    </div>
                                    <div className="info-row">
                                        <label>Last Name</label>
                                        <span>{user.lastName || user.lastname || '-'}</span>
                                    </div>
                                    <div className="info-row">
                                        <label>Role</label>
                                        <span className="role-badge">{user.role || '-'}</span>
                                    </div>
                                </div>
                            </div>

                            <div className="profile-section">
                                <h3>Contact Information</h3>
                                <div className="info-list">
                                    <div className="info-row">
                                        <label>Email</label>
                                        <span>{user.email || '-'}</span>
                                    </div>
                                    <div className="info-row">
                                        <label>Mobile</label>
                                        <span>{user.mobile || user.mobile_number || '-'}</span>
                                    </div>
                                </div>
                            </div>

                            <div className="profile-section">
                                <h3>Account Information</h3>
                                <div className="info-list">
                                    <div className="info-row">
                                        <label>Account Status</label>
                                        <span className={`status-badge ${statusInfo.className}`}>
                                            {statusInfo.label}
                                        </span>
                                    </div>
                                    <div className="info-row">
                                        <label>User Since</label>
                                        <span>{user.createdAt ? new Date(user.createdAt).toLocaleDateString('en-US', {
                                            year: 'numeric',
                                            month: 'long',
                                            day: 'numeric'
                                        }) : '-'}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </Layout>
    );
};

export default Profile;


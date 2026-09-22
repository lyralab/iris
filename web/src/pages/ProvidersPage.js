import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import './ProvidersPage.css';

const ProvidersPage = () => {
    const [providers, setProviders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showEditModal, setShowEditModal] = useState(false);
    const [selectedProvider, setSelectedProvider] = useState(null);
    const [editData, setEditData] = useState({ priority: '', status: true });
    const [actionError, setActionError] = useState('');

    useEffect(() => {
        fetchProviders();
    }, []);

    const fetchProviders = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getProviders();
            setProviders(data.providers || []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const handleEditProvider = (provider) => {
        setSelectedProvider(provider);
        setEditData({
            priority: provider.priority,
            status: provider.enabled,
        });
        setShowEditModal(true);
    };

    const handleUpdateProvider = async (e) => {
        e.preventDefault();
        setActionError('');
        try {
            await apiService.modifyProvider({
                name: selectedProvider.name,
                priority: parseInt(editData.priority),
                status: editData.status,
            });
            setShowEditModal(false);
            setSelectedProvider(null);
            fetchProviders();
        } catch (err) {
            setActionError(err.message);
        }
    };

    const handleToggleStatus = async (provider) => {
        setActionError('');
        try {
            await apiService.modifyProvider({
                name: provider.name,
                status: !provider.enabled,
            });
            fetchProviders();
        } catch (err) {
            setActionError(err.message);
        }
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading providers...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading providers: {error}</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="providers-page">
                <div className="page-header">
                    <div>
                        <h1>Notification Providers</h1>
                        <p className="subtitle">Manage notification providers and their priorities</p>
                    </div>
                </div>

                {actionError && !showEditModal && (
                    <div className="error-message">{actionError}</div>
                )}

                <div className="providers-grid">
                    {providers.map((provider) => (
                        <div key={provider.id} className={`provider-card ${provider.enabled ? 'enabled' : 'disabled'}`}>
                            <div className="provider-header">
                                <h3>{provider.name}</h3>
                                <span className={`status-badge ${provider.enabled ? 'status-active' : 'status-disable'}`}>
                                    {provider.enabled ? 'Enabled' : 'Disabled'}
                                </span>
                            </div>
                            <div className="provider-body">
                                <p className="description">{provider.description || 'No description'}</p>
                                <div className="provider-info">
                                    <div className="info-item">
                                        <span className="label">Priority:</span>
                                        <span className="value">
                                            <span className={`priority-badge priority-${provider.priority}`}>
                                                {provider.priority}
                                            </span>
                                        </span>
                                    </div>
                                    <div className="info-item">
                                        <span className="label">ID:</span>
                                        <span className="value small">{provider.id}</span>
                                    </div>
                                </div>
                            </div>
                            <div className="provider-actions">
                                <button
                                    className="btn-secondary"
                                    onClick={() => handleEditProvider(provider)}
                                >
                                    Edit Settings
                                </button>
                                <button
                                    className={provider.enabled ? 'btn-danger' : 'btn-primary'}
                                    onClick={() => handleToggleStatus(provider)}
                                >
                                    {provider.enabled ? 'Disable' : 'Enable'}
                                </button>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Edit Provider Modal */}
                {showEditModal && selectedProvider && (
                    <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Edit Provider: {selectedProvider.name}</h2>
                            {actionError && <div className="error-message">{actionError}</div>}
                            <form onSubmit={handleUpdateProvider}>
                                <div className="form-group">
                                    <label>Priority (1-5):</label>
                                    <input
                                        type="number"
                                        min="1"
                                        max="5"
                                        value={editData.priority}
                                        onChange={(e) => setEditData({ ...editData, priority: e.target.value })}
                                        required
                                    />
                                    <small>Lower number = higher priority</small>
                                </div>
                                <div className="form-group checkbox-group">
                                    <label>
                                        <input
                                            type="checkbox"
                                            checked={editData.status}
                                            onChange={(e) => setEditData({ ...editData, status: e.target.checked })}
                                        />
                                        <span>Enable Provider</span>
                                    </label>
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">Update Provider</button>
                                    <button type="button" className="btn-secondary" onClick={() => setShowEditModal(false)}>
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}
            </div>
        </Layout>
    );
};

export default ProvidersPage;


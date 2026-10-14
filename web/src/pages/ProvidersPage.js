import React, { useState, useEffect, useCallback } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import './ProvidersPage.css';

const MESSAGE_LIMITS = [10, 25, 50];
const MESSAGE_LIMIT_KEY = 'iris.providers.messagesLimit';
const MESSAGE_STATUS_CLASS = { Delivered: 'status-active', Sent: 'status-new', Failed: 'status-disable' };

function readMessageLimit() {
    try {
        const n = Number(localStorage.getItem(MESSAGE_LIMIT_KEY));
        return MESSAGE_LIMITS.includes(n) ? n : MESSAGE_LIMITS[0];
    } catch {
        return MESSAGE_LIMITS[0];
    }
}

const ProviderMessagesModal = ({ provider, onClose }) => {
    const [limit, setLimit] = useState(readMessageLimit);
    const [messages, setMessages] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const load = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getProviderMessages(provider.name, limit);
            setMessages(data.messages || []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }, [provider.name, limit]);

    useEffect(() => {
        load();
    }, [load]);

    const changeLimit = (value) => {
        setLimit(value);
        try {
            localStorage.setItem(MESSAGE_LIMIT_KEY, String(value));
        } catch {
            // ignore unavailable storage
        }
    };

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal messages-modal" onClick={(e) => e.stopPropagation()}>
                <div className="messages-header">
                    <h2>Latest Messages: {provider.name}</h2>
                    <div className="messages-controls">
                        <select
                            value={limit}
                            onChange={(e) => changeLimit(Number(e.target.value))}
                            aria-label="Number of messages"
                        >
                            {MESSAGE_LIMITS.map((n) => (
                                <option key={n} value={n}>Last {n}</option>
                            ))}
                        </select>
                        <button className="btn-secondary" onClick={load} disabled={loading}>
                            {loading ? 'Loading...' : 'Refresh'}
                        </button>
                    </div>
                </div>
                {error && <div className="error-message">{error}</div>}
                {!error && !loading && messages.length === 0 && (
                    <p className="subtitle">No messages sent through this provider yet.</p>
                )}
                {messages.length > 0 && (
                    <div className="alerts-table-container">
                        <table className="alerts-table">
                            <thead>
                                <tr>
                                    <th>Time</th>
                                    <th>Receptor</th>
                                    <th>User</th>
                                    <th>Status</th>
                                    <th>Response</th>
                                    <th>Message</th>
                                </tr>
                            </thead>
                            <tbody>
                                {messages.map((m, i) => (
                                    <tr key={`${m.created_at}-${m.receptor}-${i}`}>
                                        <td className="nowrap">{new Date(m.created_at).toLocaleString()}</td>
                                        <td>{m.receptor}</td>
                                        <td>{m.user_name || '—'}</td>
                                        <td>
                                            <span className={`status-badge ${MESSAGE_STATUS_CLASS[m.status] || 'status-unknown'}`}>
                                                {m.status || 'Unknown'}
                                            </span>
                                        </td>
                                        <td className="message-response">{m.response}</td>
                                        <td className="alert-description" title={m.message}>{m.message}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
                <div className="modal-actions">
                    <button type="button" className="btn-secondary" onClick={onClose}>Close</button>
                </div>
            </div>
        </div>
    );
};

const ProvidersPage = () => {
    const [providers, setProviders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showEditModal, setShowEditModal] = useState(false);
    const [selectedProvider, setSelectedProvider] = useState(null);
    const [editData, setEditData] = useState({ priority: '', status: true });
    const [actionError, setActionError] = useState('');
    const [messagesProvider, setMessagesProvider] = useState(null);

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

    useEffect(() => {
        fetchProviders();
    }, []);

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
                                    className="btn-secondary"
                                    onClick={() => setMessagesProvider(provider)}
                                >
                                    Messages
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

                {messagesProvider && (
                    <ProviderMessagesModal
                        provider={messagesProvider}
                        onClose={() => setMessagesProvider(null)}
                    />
                )}

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


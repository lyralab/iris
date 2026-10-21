import React, { useState, useEffect, useCallback } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import EmptyState from '../components/EmptyState';
import Icon from '../components/Icon';
import Modal from '../components/Modal';
import { formatDateTime } from '../utils/format';
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
        <Modal title={`Messages · ${provider.name}`} subtitle="Latest notifications sent through this provider" size="modal-xl" onClose={onClose}>
            <div className="modal-body">
                <div className="messages-controls">
                    <select
                        className="select-input select-sm"
                        value={limit}
                        onChange={(e) => changeLimit(Number(e.target.value))}
                        aria-label="Number of messages"
                    >
                        {MESSAGE_LIMITS.map((n) => (
                            <option key={n} value={n}>Last {n}</option>
                        ))}
                    </select>
                    <button className="btn-secondary btn-sm" onClick={load} disabled={loading}>
                        <Icon name="refresh" size={14} className={loading ? 'spin' : ''} />
                        Refresh
                    </button>
                </div>
                {error && <div className="error-message"><Icon name="alert" />{error}</div>}
                {!error && !loading && messages.length === 0 && (
                    <EmptyState icon="message" title="No messages yet">Nothing has been sent through this provider.</EmptyState>
                )}
                {messages.length > 0 && (
                    <div className="alerts-table-container messages-table">
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
                                        <td className="nowrap cell-muted tabular">{formatDateTime(m.created_at)}</td>
                                        <td className="mono">{m.receptor}</td>
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
            </div>
            <div className="modal-footer">
                <button type="button" className="btn-secondary" onClick={onClose}>Close</button>
            </div>
        </Modal>
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

    const sorted = [...providers].sort((x, y) => (x.priority ?? 99) - (y.priority ?? 99));
    const enabledCount = providers.filter((p) => p.enabled).length;

    let content;
    if (loading) {
        content = <div className="loading">Loading providers…</div>;
    } else if (error) {
        content = <div className="error-message"><Icon name="alert" />Error loading providers: {error}</div>;
    } else if (providers.length === 0) {
        content = <div className="card"><EmptyState icon="send" title="No providers configured" /></div>;
    } else {
        content = (
            <div className="providers-grid">
                {sorted.map((provider) => (
                    <article key={provider.id} className={`card provider-card ${provider.enabled ? 'enabled' : 'disabled'}`}>
                        <div className="provider-header">
                            <span className="provider-icon"><Icon name="send" size={17} /></span>
                            <div className="provider-title">
                                <h3>{provider.name}</h3>
                                <span className={`status-badge ${provider.enabled ? 'status-active' : 'status-unknown'}`}>
                                    {provider.enabled ? 'Enabled' : 'Disabled'}
                                </span>
                            </div>
                            <label className="switch" title={provider.enabled ? 'Disable provider' : 'Enable provider'}>
                                <input
                                    type="checkbox"
                                    checked={!!provider.enabled}
                                    onChange={() => handleToggleStatus(provider)}
                                    aria-label={`${provider.enabled ? 'Disable' : 'Enable'} ${provider.name}`}
                                />
                                <span className="switch-track" />
                            </label>
                        </div>
                        <p className="description">{provider.description || 'No description'}</p>
                        <dl className="provider-info">
                            <div>
                                <dt>Priority</dt>
                                <dd><span className={`priority-badge priority-${provider.priority}`}>P{provider.priority}</span></dd>
                            </div>
                            <div>
                                <dt>ID</dt>
                                <dd className="mono provider-id" title={provider.id}>{provider.id}</dd>
                            </div>
                        </dl>
                        <div className="provider-actions">
                            <button className="btn-ghost btn-sm" onClick={() => setMessagesProvider(provider)}>
                                <Icon name="message" size={14} />
                                Messages
                            </button>
                            <button className="btn-secondary btn-sm" onClick={() => handleEditProvider(provider)}>
                                <Icon name="sliders" size={14} />
                                Settings
                            </button>
                        </div>
                    </article>
                ))}
            </div>
        );
    }

    return (
        <Layout>
            <div className="page providers-page">
                <div className="page-header">
                    <div>
                        <h1>Providers</h1>
                        <p className="subtitle">
                            Notification channels, tried in priority order (P1 first)
                            {!loading && !error && providers.length > 0 && <> · {enabledCount} of {providers.length} enabled</>}
                        </p>
                    </div>
                </div>

                {actionError && !showEditModal && (
                    <div className="error-message"><Icon name="alert" />{actionError}</div>
                )}

                {content}

                {messagesProvider && (
                    <ProviderMessagesModal
                        provider={messagesProvider}
                        onClose={() => setMessagesProvider(null)}
                    />
                )}

                {showEditModal && selectedProvider && (
                    <Modal title="Provider settings" subtitle={selectedProvider.name} onClose={() => setShowEditModal(false)}>
                        <form onSubmit={handleUpdateProvider}>
                            <div className="modal-body">
                                {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}
                                <div className="form-group">
                                    <label htmlFor="provider-priority">Priority</label>
                                    <input
                                        id="provider-priority"
                                        type="number"
                                        min="1"
                                        max="5"
                                        value={editData.priority}
                                        onChange={(e) => setEditData({ ...editData, priority: e.target.value })}
                                        required
                                    />
                                    <small>1–5. Lower number = tried first.</small>
                                </div>
                                <label className="switch">
                                    <input
                                        type="checkbox"
                                        checked={editData.status}
                                        onChange={(e) => setEditData({ ...editData, status: e.target.checked })}
                                    />
                                    <span className="switch-track" />
                                    <span>Provider enabled</span>
                                </label>
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn-secondary" onClick={() => setShowEditModal(false)}>Cancel</button>
                                <button type="submit" className="btn-primary">Save</button>
                            </div>
                        </form>
                    </Modal>
                )}
            </div>
        </Layout>
    );
};

export default ProvidersPage;

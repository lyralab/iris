import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import ConfirmDialog from '../components/ConfirmDialog';
import EmptyState from '../components/EmptyState';
import Icon from '../components/Icon';
import Modal from '../components/Modal';
import { initials, userStatusInfo } from '../utils/format';
import './UserManagement.css';

const CHANNELS = [
    { key: 'mobile', label: 'SMS' },
    { key: 'email', label: 'Email' },
    { key: 'telegram_id', label: 'Telegram' },
    { key: 'mattermost_id', label: 'Mattermost' },
    { key: 'bale_id', label: 'Bale' },
];

// Fields shared by the add and edit forms. `isNew` adds the password fields and unlocks username.
const UserFields = ({ value, onChange, isNew }) => {
    const set = (field) => (e) => onChange({ ...value, [field]: e.target.value });
    return (
        <>
            <div className="form-section">
                <h3 className="form-section-title">Identity</h3>
                <div className="form-group">
                    <label htmlFor="user-username">Username</label>
                    <input
                        id="user-username"
                        type="text"
                        value={value.username}
                        onChange={set('username')}
                        required
                        minLength="3"
                        disabled={!isNew}
                    />
                </div>
                <div className="form-row">
                    <div className="form-group">
                        <label htmlFor="user-firstname">First name</label>
                        <input id="user-firstname" type="text" value={value.firstname || ''} onChange={set('firstname')} />
                    </div>
                    <div className="form-group">
                        <label htmlFor="user-lastname">Last name</label>
                        <input id="user-lastname" type="text" value={value.lastname || ''} onChange={set('lastname')} />
                    </div>
                </div>
            </div>

            <div className="form-section">
                <h3 className="form-section-title">Notification channels</h3>
                <div className="form-row">
                    <div className="form-group">
                        <label htmlFor="user-email">Email</label>
                        <input id="user-email" type="email" value={value.email || ''} onChange={set('email')} />
                    </div>
                    <div className="form-group">
                        <label htmlFor="user-mobile">Mobile number</label>
                        <input
                            id="user-mobile"
                            type="text"
                            value={value.mobile_number || ''}
                            onChange={set('mobile_number')}
                            pattern="[0-9]{11}"
                            placeholder="09xxxxxxxxx"
                        />
                        <small>11 digits</small>
                    </div>
                </div>
                <div className="form-row">
                    <div className="form-group">
                        <label htmlFor="user-telegram">Telegram ID</label>
                        <input
                            id="user-telegram"
                            type="text"
                            value={value.telegram_id || ''}
                            onChange={set('telegram_id')}
                            pattern="-?[0-9]+"
                            maxLength="50"
                            placeholder="User or channel ID"
                        />
                        <small>Channel IDs are negative, e.g. -1001234567890</small>
                    </div>
                    <div className="form-group">
                        <label htmlFor="user-bale">Bale ID</label>
                        <input
                            id="user-bale"
                            type="text"
                            value={value.bale_id || ''}
                            onChange={set('bale_id')}
                            pattern="-?[0-9]+"
                            maxLength="50"
                            placeholder="User or channel ID"
                        />
                    </div>
                </div>
                <div className="form-group">
                    <label htmlFor="user-mattermost">Mattermost ID</label>
                    <input
                        id="user-mattermost"
                        type="text"
                        value={value.mattermost_id || ''}
                        onChange={set('mattermost_id')}
                        maxLength="50"
                        placeholder="Mattermost user or channel ID"
                    />
                </div>
            </div>

            {isNew && (
                <div className="form-section">
                    <h3 className="form-section-title">Password</h3>
                    <div className="form-row">
                        <div className="form-group">
                            <label htmlFor="user-password">Password</label>
                            <input
                                id="user-password"
                                type="password"
                                value={value.password}
                                onChange={set('password')}
                                required
                                minLength="8"
                                autoComplete="new-password"
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="user-password2">Confirm password</label>
                            <input
                                id="user-password2"
                                type="password"
                                value={value['confirm-password']}
                                onChange={set('confirm-password')}
                                required
                                autoComplete="new-password"
                            />
                        </div>
                    </div>
                    <small className="form-hint">Min 8 characters, with uppercase, lowercase, a digit and a special character.</small>
                </div>
            )}
        </>
    );
};

const UserManagement = () => {
    const [users, setUsers] = useState([]);
    const [groups, setGroups] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showAddModal, setShowAddModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [selectedUser, setSelectedUser] = useState(null);
    const [showGroupModal, setShowGroupModal] = useState(false);
    const [selectedUserForGroup, setSelectedUserForGroup] = useState(null);
    const [newUser, setNewUser] = useState({
        username: '',
        firstname: '',
        lastname: '',
        email: '',
        mobile_number: '',
        telegram_id: '',
        mattermost_id: '',
        bale_id: '',
        password: '',
        'confirm-password': ''
    });
    const [confirm, setConfirm] = useState(null);
    const [actionError, setActionError] = useState('');
    const [query, setQuery] = useState('');

    const fetchUsers = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getUsers();
            setUsers(data.users || []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const fetchGroups = async () => {
        try {
            const data = await apiService.getGroups();
            setGroups(data.data || []);
        } catch (e) {
            console.log('Error fetching groups:', e);
        }
    };

    useEffect(() => {
        fetchUsers();
        fetchGroups();
    }, []);

    const handleAddUser = async (e) => {
        e.preventDefault();
        setActionError('');

        if (newUser.password !== newUser['confirm-password']) {
            setActionError('Passwords do not match');
            return;
        }

        try {
            await apiService.addUser(newUser);
            setShowAddModal(false);
            setNewUser({
                username: '',
                firstname: '',
                lastname: '',
                email: '',
                mobile_number: '',
                telegram_id: '',
                mattermost_id: '',
                bale_id: '',
                password: '',
                'confirm-password': ''
            });
            fetchUsers();
        } catch (err) {
            setActionError(err.message);
        }
    };

    const handleUpdateUser = async (e) => {
        e.preventDefault();
        setActionError('');
        try {
            await apiService.updateUser({
                username: selectedUser.username,
                firstname: selectedUser.firstname,
                lastname: selectedUser.lastname,
                email: selectedUser.email,
                mobile_number: selectedUser.mobile_number,
                telegram_id: selectedUser.telegram_id,
                mattermost_id: selectedUser.mattermost_id,
                bale_id: selectedUser.bale_id,
            });
            setShowEditModal(false);
            setSelectedUser(null);
            fetchUsers();
        } catch (err) {
            setActionError(err.message);
        }
    };

    const handleVerifyUser = (username) => {
        setActionError('');
        setConfirm({
            title: 'Verify user',
            message: `Verify ${username}?`,
                        confirmLabel: 'Verify',
                        danger: false,
                        onConfirm: async () => {
                setConfirm(null);
                try {
                    await apiService.verifyUser(username);
                    fetchUsers();
                } catch (err) {
                    setActionError(err.message);
                }
            },
        });
    };

    const handleAddUserToGroup = async (groupId) => {
        setActionError('');
        try {
            await apiService.addUserToGroup(groupId, selectedUserForGroup.id);
            setShowGroupModal(false);
            setSelectedUserForGroup(null);
        } catch (err) {
            setActionError(err.message);
        }
    };

    const needle = query.trim().toLowerCase();
    const visibleUsers = users.filter((user) => {
        if (!needle) return true;
        const name = `${user.firstName || ''} ${user.lastName || ''}`;
        return [user.username, name, user.email].join(' ').toLowerCase().includes(needle);
    });
    const pendingCount = users.filter((u) => String(u.status).toLowerCase() === 'new').length;

    const openEdit = (user) => {
        setActionError('');
        setSelectedUser({
            username: user.username,
            firstname: user.firstName || '',
            lastname: user.lastName || '',
            email: user.email || '',
            mobile_number: user.mobile || '',
            telegram_id: user.telegram_id || '',
            mattermost_id: user.mattermost_id || '',
            bale_id: user.bale_id || '',
        });
        setShowEditModal(true);
    };

    let content;
    if (loading) {
        content = <div className="loading">Loading users…</div>;
    } else if (error) {
        content = <div className="card-body"><div className="error-message"><Icon name="alert" />Error loading users: {error}</div></div>;
    } else if (visibleUsers.length === 0) {
        content = users.length === 0
            ? <EmptyState icon="users" title="No users yet">Add the first person who should receive alerts.</EmptyState>
            : <EmptyState icon="search" title="No users match">Try a different name or email.</EmptyState>;
    } else {
        content = (
            <div className="users-table-container">
                <table className="users-table">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Contact</th>
                            <th>Channels</th>
                            <th>Status</th>
                            <th className="cell-actions"><span className="visually-hidden">Actions</span></th>
                        </tr>
                    </thead>
                    <tbody>
                        {visibleUsers.map((user) => {
                            const statusInfo = userStatusInfo(user.status);
                            const fullName = [user.firstName, user.lastName].filter(Boolean).join(' ');
                            return (
                                <tr key={user.id}>
                                    <td>
                                        <div className="user-name-cell">
                                            <span className="avatar">{initials(user.firstName, user.lastName) !== '?' ? initials(user.firstName, user.lastName) : initials(user.username)}</span>
                                            <div>
                                                <strong>{user.username}</strong>
                                                {fullName && <span>{fullName}</span>}
                                            </div>
                                        </div>
                                    </td>
                                    <td>
                                        <div className="contact-cell">
                                            <span>{user.email || <span className="cell-muted">No email</span>}</span>
                                            {user.mobile && <span className="cell-muted tabular">{user.mobile}</span>}
                                        </div>
                                    </td>
                                    <td>
                                        <div className="channel-list">
                                            {CHANNELS.filter((c) => user[c.key]).map((c) => (
                                                <span key={c.key} className="badge badge-outline">{c.label}</span>
                                            ))}
                                            {!CHANNELS.some((c) => user[c.key]) && <span className="cell-muted">None</span>}
                                        </div>
                                    </td>
                                    <td>
                                        <span className={`status-badge ${statusInfo.className}`}>{statusInfo.label}</span>
                                    </td>
                                    <td className="cell-actions">
                                        <div className="actions-cell">
                                            {String(user.status).toLowerCase() === 'new' && (
                                                <button className="btn-verify btn-sm" onClick={() => handleVerifyUser(user.username)}>
                                                    <Icon name="check" size={14} />
                                                    Verify
                                                </button>
                                            )}
                                            <button className="btn-ghost btn-sm" onClick={() => openEdit(user)}>
                                                <Icon name="edit" size={14} />
                                                Edit
                                            </button>
                                            <button
                                                className="btn-ghost btn-sm"
                                                onClick={() => {
                                                    setActionError('');
                                                    setSelectedUserForGroup(user);
                                                    setShowGroupModal(true);
                                                }}
                                            >
                                                <Icon name="group" size={14} />
                                                Groups
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        );
    }

    return (
        <Layout>
            <div className="page user-management">
                <div className="page-header">
                    <div>
                        <h1>Users</h1>
                        <p className="subtitle">
                            People who receive alert notifications
                            {pendingCount > 0 && <> · <strong className="pending-note">{pendingCount} awaiting verification</strong></>}
                        </p>
                    </div>
                    <div className="header-actions">
                        <div className="input-with-icon">
                            <Icon name="search" />
                            <input
                                className="search-input"
                                type="search"
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                placeholder="Search name or email"
                                aria-label="Search users"
                            />
                        </div>
                        <button className="btn-primary" onClick={() => { setActionError(''); setShowAddModal(true); }}>
                            <Icon name="userPlus" />
                            Add user
                        </button>
                    </div>
                </div>

                {actionError && !showAddModal && !showEditModal && !showGroupModal && (
                    <div className="error-message"><Icon name="alert" />{actionError}</div>
                )}

                <section className="card table-card">{content}</section>

                {showAddModal && (
                    <Modal title="Add user" subtitle="New users start unverified until an admin verifies them." onClose={() => setShowAddModal(false)}>
                        <form onSubmit={handleAddUser}>
                            <div className="modal-body">
                                {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}
                                <UserFields value={newUser} onChange={setNewUser} isNew />
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)}>Cancel</button>
                                <button type="submit" className="btn-primary">Add user</button>
                            </div>
                        </form>
                    </Modal>
                )}

                {showEditModal && selectedUser && (
                    <Modal title="Edit user" subtitle={selectedUser.username} onClose={() => setShowEditModal(false)}>
                        <form onSubmit={handleUpdateUser}>
                            <div className="modal-body">
                                {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}
                                <UserFields value={selectedUser} onChange={setSelectedUser} />
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn-secondary" onClick={() => setShowEditModal(false)}>Cancel</button>
                                <button type="submit" className="btn-primary">Save changes</button>
                            </div>
                        </form>
                    </Modal>
                )}

                {showGroupModal && selectedUserForGroup && (
                    <Modal
                        title="Add to group"
                        subtitle={`Choose a group for ${selectedUserForGroup.username}`}
                        onClose={() => setShowGroupModal(false)}
                    >
                        <div className="modal-body">
                            {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}
                            {groups.length === 0 ? (
                                <EmptyState icon="group" title="No groups yet">Create a group on the Groups page first.</EmptyState>
                            ) : (
                                <ul className="groups-list">
                                    {groups.map((group) => (
                                        <li key={group.group_id} className="group-item">
                                            <span className="group-item-icon"><Icon name="group" /></span>
                                            <span className="group-item-text">
                                                <strong>{group.group_name}</strong>
                                                {group.description && <span>{group.description}</span>}
                                            </span>
                                            <button className="btn-secondary btn-sm" onClick={() => handleAddUserToGroup(group.group_id)}>
                                                <Icon name="plus" size={14} />
                                                Add
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={() => setShowGroupModal(false)}>Close</button>
                        </div>
                    </Modal>
                )}

                {confirm && (
                    <ConfirmDialog
                        title={confirm.title}
                        message={confirm.message}
                        confirmLabel={confirm.confirmLabel}
                        danger={confirm.danger}
                        onConfirm={confirm.onConfirm}
                        onCancel={() => setConfirm(null)}
                    />
                )}
            </div>
        </Layout>
    );
};

export default UserManagement;

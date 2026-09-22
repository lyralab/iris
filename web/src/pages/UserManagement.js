import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import ConfirmDialog from '../components/ConfirmDialog';
import './UserManagement.css';

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

    // Function to get status display info
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

    useEffect(() => {
        fetchUsers();
        fetchGroups();
    }, []);

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

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading users...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading users: {error}</div>
            </Layout>
        );
    }

    const needle = query.trim().toLowerCase();
    const visibleUsers = users.filter((user) => {
        if (!needle) return true;
        const name = `${user.firstName || ''} ${user.lastName || ''}`;
        return [user.username, name, user.email].join(' ').toLowerCase().includes(needle);
    });

    return (
        <Layout>
            <div className="user-management">
                <div className="page-header">
                    <h1>User Management</h1>
                    <div className="header-actions">
                        <input
                            className="search-input"
                            type="search"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            placeholder="Search name or email"
                        />
                        <button className="btn-primary" onClick={() => setShowAddModal(true)}>
                            + Add User
                        </button>
                    </div>
                </div>

                {actionError && !showAddModal && !showEditModal && !showGroupModal && (
                    <div className="error-message">{actionError}</div>
                )}

                {visibleUsers.length === 0 ? (
                    <p className="empty">{users.length === 0 ? 'No users' : 'No users match'}</p>
                ) : (
                <div className="users-table-container">
                    <table className="users-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Email</th>
                                <th>Mobile</th>
                                <th>Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {visibleUsers.map((user) => (
                                <tr key={user.id}>
                                    <td>
                                        <div className="user-name-cell">
                                            <strong>{user.username}</strong>
                                            {(user.firstName || user.lastName) && (
                                                <span>{[user.firstName, user.lastName].filter(Boolean).join(' ')}</span>
                                            )}
                                        </div>
                                    </td>
                                    <td>{user.email || '-'}</td>
                                    <td>{user.mobile || '-'}</td>
                                    <td>
                                        {(() => {
                                            const statusInfo = getStatusInfo(user.status);
                                            return (
                                                <span className={`status-badge ${statusInfo.className}`}>
                                                    {statusInfo.label}
                                                </span>
                                            );
                                        })()}
                                    </td>
                                    <td className="actions-cell">
                                        <button
                                            className="btn-edit"
                                            onClick={() => {
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
                                            }}
                                        >
                                            Edit
                                        </button>
                                        <button
                                            className="btn-group"
                                            onClick={() => {
                                                setSelectedUserForGroup(user);
                                                setShowGroupModal(true);
                                            }}
                                        >
                                            Groups
                                        </button>
                                        {(user.status === 'New' || user.status === 'new') && (
                                            <button
                                                className="btn-verify"
                                                onClick={() => handleVerifyUser(user.username)}
                                            >
                                                Verify
                                            </button>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
                )}

                {/* Add User Modal */}
                {showAddModal && (
                    <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Add New User</h2>
                            {actionError && <div className="error-message">{actionError}</div>}
                            <form onSubmit={handleAddUser}>
                                <div className="form-group">
                                    <label>Username:</label>
                                    <input
                                        type="text"
                                        value={newUser.username}
                                        onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                                        required
                                        minLength="3"
                                    />
                                </div>
                                <div className="form-row">
                                    <div className="form-group">
                                        <label>First Name:</label>
                                        <input
                                            type="text"
                                            value={newUser.firstname}
                                            onChange={(e) => setNewUser({ ...newUser, firstname: e.target.value })}
                                        />
                                    </div>
                                    <div className="form-group">
                                        <label>Last Name:</label>
                                        <input
                                            type="text"
                                            value={newUser.lastname}
                                            onChange={(e) => setNewUser({ ...newUser, lastname: e.target.value })}
                                        />
                                    </div>
                                </div>
                                <div className="form-group">
                                    <label>Email:</label>
                                    <input
                                        type="email"
                                        value={newUser.email}
                                        onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Mobile Number:</label>
                                    <input
                                        type="text"
                                        value={newUser.mobile_number}
                                        onChange={(e) => setNewUser({ ...newUser, mobile_number: e.target.value })}
                                        pattern="[0-9]{11}"
                                        placeholder="09xxxxxxxxx"
                                    />
                                    <small>11 digits required</small>
                                </div>
                                <div className="form-group">
                                    <label>Telegram ID:</label>
                                    <input
                                        type="text"
                                        value={newUser.telegram_id}
                                        onChange={(e) => setNewUser({ ...newUser, telegram_id: e.target.value })}
                                        pattern="-?[0-9]+"
                                        maxLength="50"
                                        placeholder="User or channel ID"
                                    />
                                    <small>Channel IDs are negative, for example -1001234567890</small>
                                </div>
                                <div className="form-group">
                                    <label>Mattermost ID:</label>
                                    <input
                                        type="text"
                                        value={newUser.mattermost_id}
                                        onChange={(e) => setNewUser({ ...newUser, mattermost_id: e.target.value })}
                                        maxLength="50"
                                        placeholder="Mattermost user or channel ID"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Bale ID:</label>
                                    <input
                                        type="text"
                                        value={newUser.bale_id}
                                        onChange={(e) => setNewUser({ ...newUser, bale_id: e.target.value })}
                                        pattern="-?[0-9]+"
                                        maxLength="50"
                                        placeholder="User or channel ID"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Password:</label>
                                    <input
                                        type="password"
                                        value={newUser.password}
                                        onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                                        required
                                        minLength="8"
                                    />
                                    <small>Min 8 characters, must include uppercase, lowercase, digit, and special character</small>
                                </div>
                                <div className="form-group">
                                    <label>Confirm Password:</label>
                                    <input
                                        type="password"
                                        value={newUser['confirm-password']}
                                        onChange={(e) => setNewUser({ ...newUser, 'confirm-password': e.target.value })}
                                        required
                                    />
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">Add User</button>
                                    <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)}>
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}

                {/* Edit User Modal */}
                {showEditModal && selectedUser && (
                    <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Edit User</h2>
                            {actionError && <div className="error-message">{actionError}</div>}
                            <form onSubmit={handleUpdateUser}>
                                <div className="form-group">
                                    <label>Username:</label>
                                    <input
                                        type="text"
                                        value={selectedUser.username}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, username: e.target.value })}
                                        required
                                        disabled
                                    />
                                </div>
                                <div className="form-row">
                                    <div className="form-group">
                                        <label>First Name:</label>
                                        <input
                                            type="text"
                                            value={selectedUser.firstname || ''}
                                            onChange={(e) => setSelectedUser({ ...selectedUser, firstname: e.target.value })}
                                        />
                                    </div>
                                    <div className="form-group">
                                        <label>Last Name:</label>
                                        <input
                                            type="text"
                                            value={selectedUser.lastname || ''}
                                            onChange={(e) => setSelectedUser({ ...selectedUser, lastname: e.target.value })}
                                        />
                                    </div>
                                </div>
                                <div className="form-group">
                                    <label>Email:</label>
                                    <input
                                        type="email"
                                        value={selectedUser.email || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, email: e.target.value })}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Mobile Number:</label>
                                    <input
                                        type="text"
                                        value={selectedUser.mobile_number || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, mobile_number: e.target.value })}
                                        pattern="[0-9]{11}"
                                        placeholder="09xxxxxxxxx"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Telegram ID:</label>
                                    <input
                                        type="text"
                                        value={selectedUser.telegram_id || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, telegram_id: e.target.value })}
                                        pattern="-?[0-9]+"
                                        maxLength="50"
                                        placeholder="User or channel ID"
                                    />
                                    <small>Channel IDs are negative, for example -1001234567890</small>
                                </div>
                                <div className="form-group">
                                    <label>Mattermost ID:</label>
                                    <input
                                        type="text"
                                        value={selectedUser.mattermost_id || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, mattermost_id: e.target.value })}
                                        maxLength="50"
                                        placeholder="Mattermost user or channel ID"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Bale ID:</label>
                                    <input
                                        type="text"
                                        value={selectedUser.bale_id || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, bale_id: e.target.value })}
                                        pattern="-?[0-9]+"
                                        maxLength="50"
                                        placeholder="User or channel ID"
                                    />
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">Update User</button>
                                    <button type="button" className="btn-secondary" onClick={() => setShowEditModal(false)}>
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}

                {/* Add to Group Modal */}
                {showGroupModal && selectedUserForGroup && (
                    <div className="modal-overlay" onClick={() => setShowGroupModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Add {selectedUserForGroup.username} to Group</h2>
                            {actionError && <div className="error-message">{actionError}</div>}
                            <div className="groups-list">
                                {groups.map((group) => (
                                    <div key={group.group_id} className="group-item">
                                        <span>{group.group_name}</span>
                                        <button
                                            className="btn-primary"
                                            onClick={() => handleAddUserToGroup(group.group_id)}
                                        >
                                            Add to Group
                                        </button>
                                    </div>
                                ))}
                            </div>
                            <div className="modal-actions">
                                <button className="btn-secondary" onClick={() => setShowGroupModal(false)}>
                                    Close
                                </button>
                            </div>
                        </div>
                    </div>
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


import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import ConfirmDialog from '../components/ConfirmDialog';
import './GroupManagement.css';

const GroupManagement = () => {
    const [groups, setGroups] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showAddModal, setShowAddModal] = useState(false);
    const [showMembersModal, setShowMembersModal] = useState(false);
    const [selectedGroup, setSelectedGroup] = useState(null);
    const [groupMembers, setGroupMembers] = useState([]);
    const [allUsers, setAllUsers] = useState([]);
    const [selectedUserId, setSelectedUserId] = useState('');
    const [showAddUserSection, setShowAddUserSection] = useState(false);
    const [newGroup, setNewGroup] = useState({
        name: '',
        description: ''
    });
    const [confirm, setConfirm] = useState(null);
    const [actionError, setActionError] = useState('');

    const fetchGroups = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getGroups();
            setGroups(data.data || []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const fetchAllUsers = async () => {
        try {
            const data = await apiService.getUsers();
            setAllUsers(data.users || []);
        } catch (e) {
            console.log('Error fetching users:', e);
            setAllUsers([]);
        }
    };

    useEffect(() => {
        fetchGroups();
        fetchAllUsers();
    }, []);

    const fetchGroupMembers = async (groupId) => {
        try {
            const data = await apiService.getGroupUsers(groupId);
            setGroupMembers(data.users || []);
        } catch (e) {
            console.log('Error fetching group members:', e);
            setGroupMembers([]);
        }
    };

    const askConfirm = (title, message, confirmLabel, action) => {
        setActionError('');
        setConfirm({
            title,
            message,
            confirmLabel,
            onConfirm: async () => {
                setConfirm(null);
                try {
                    await action();
                } catch (err) {
                    setActionError(err.message);
                }
            },
        });
    };

    const handleAddGroup = async (e) => {
        e.preventDefault();
        setActionError('');
        try {
            await apiService.createGroup(newGroup);
            setShowAddModal(false);
            setNewGroup({ name: '', description: '' });
            fetchGroups();
        } catch (err) {
            setActionError(err.message);
        }
    };

    const handleDeleteGroup = (group) => {
        askConfirm('Delete group', `Delete ${group.group_name}?`, 'Delete', async () => {
            await apiService.deleteGroup({ id: group.group_id });
            fetchGroups();
        });
    };

    const handleViewMembers = async (group) => {
        setSelectedGroup(group);
        await fetchGroupMembers(group.group_id);
        setActionError('');
        setShowMembersModal(true);
        setShowAddUserSection(false);
        setSelectedUserId('');
    };

    const handleAddUserToGroup = async () => {
        if (!selectedUserId) return;

        setActionError('');
        try {
            await apiService.addUserToGroup(selectedGroup.group_id, selectedUserId);
            await fetchGroupMembers(selectedGroup.group_id);
            setSelectedUserId('');
            setShowAddUserSection(false);
        } catch (err) {
            setActionError(err.message);
        }
    };

    const handleRemoveUserFromGroup = (userId, username) => {
        askConfirm(
            'Remove member',
            `Remove ${username} from ${selectedGroup.group_name}?`,
            'Remove',
            async () => {
                await apiService.removeUserFromGroup(selectedGroup.group_id, userId);
                await fetchGroupMembers(selectedGroup.group_id);
            },
        );
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading groups...</div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout>
                <div className="error-message">Error loading groups: {error}</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="group-management">
                <div className="page-header">
                    <h1>Group Management</h1>
                    <button className="btn-primary" onClick={() => { setActionError(''); setShowAddModal(true); }}>
                        + Add Group
                    </button>
                </div>

                {actionError && !showAddModal && !showMembersModal && (
                    <div className="error-message">{actionError}</div>
                )}

                {groups.length === 0 ? (
                    <p className="empty">No groups yet</p>
                ) : (
                    <div className="groups-grid">
                        {groups.map((group) => (
                            <div key={group.group_id} className="group-card">
                                <div className="group-body">
                                    <h3>{group.group_name}</h3>
                                    <p className="group-description">{group.description || 'No description'}</p>
                                </div>
                                <div className="group-actions">
                                    <button
                                        className="btn-primary"
                                        onClick={() => handleViewMembers(group)}
                                    >
                                        Members
                                    </button>
                                    <button
                                        className="btn-danger"
                                        onClick={() => handleDeleteGroup(group)}
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}

                {/* Add Group Modal */}
                {showAddModal && (
                    <div className="modal-overlay" onClick={() => { setShowAddModal(false); setActionError(''); }}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Add New Group</h2>
                            {actionError && <div className="error-message">{actionError}</div>}
                            <form onSubmit={handleAddGroup}>
                                <div className="form-group">
                                    <label>Group Name:</label>
                                    <input
                                        type="text"
                                        value={newGroup.name}
                                        onChange={(e) => setNewGroup({ ...newGroup, name: e.target.value })}
                                        required
                                        minLength="3"
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Description:</label>
                                    <textarea
                                        value={newGroup.description}
                                        onChange={(e) => setNewGroup({ ...newGroup, description: e.target.value })}
                                        rows="4"
                                    />
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">Add Group</button>
                                    <button type="button" className="btn-secondary" onClick={() => { setShowAddModal(false); setActionError(''); }}>
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}

                {/* Group Members Modal */}
                {showMembersModal && selectedGroup && (
                    <div className="modal-overlay" onClick={() => { setShowMembersModal(false); setActionError(''); }}>
                        <div className="modal modal-large" onClick={(e) => e.stopPropagation()}>
                            <h2>Members of {selectedGroup.group_name}</h2>
                            {actionError && <div className="error-message">{actionError}</div>}

                            {/* Add User Section */}
                            <div className="add-user-section">
                                {!showAddUserSection ? (
                                    <button
                                        className="btn-primary btn-add-user"
                                        onClick={() => setShowAddUserSection(true)}
                                    >
                                        + Add User to Group
                                    </button>
                                ) : (
                                    <div className="add-user-form">
                                        <div className="form-group">
                                            <label>Select User:</label>
                                            <select
                                                value={selectedUserId}
                                                onChange={(e) => setSelectedUserId(e.target.value)}
                                                className="user-dropdown"
                                            >
                                                <option value="">-- Select a user --</option>
                                                {allUsers
                                                    .filter(user => !groupMembers.some(member => member.id === user.id))
                                                    .map(user => (
                                                        <option key={user.id} value={user.id}>
                                                            {user.username} ({user.email || user.mobile || 'No contact'})
                                                        </option>
                                                    ))
                                                }
                                            </select>
                                        </div>
                                        <div className="add-user-actions">
                                            <button
                                                className="btn-primary"
                                                onClick={handleAddUserToGroup}
                                                disabled={!selectedUserId}
                                            >
                                                Add User
                                            </button>
                                            <button
                                                className="btn-secondary"
                                                onClick={() => {
                                                    setShowAddUserSection(false);
                                                    setSelectedUserId('');
                                                }}
                                            >
                                                Cancel
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>

                            <div className="members-list">
                                {groupMembers.length === 0 ? (
                                    <p className="no-members">No members in this group</p>
                                ) : (
                                    <table className="members-table">
                                        <thead>
                                            <tr>
                                                <th>Username</th>
                                                <th>Email</th>
                                                <th>Mobile</th>
                                                <th></th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {groupMembers.map((member) => (
                                                <tr key={member.id || member.username}>
                                                    <td>{member.username}</td>
                                                    <td>{member.email || '-'}</td>
                                                    <td>{member.mobile || '-'}</td>
                                                    <td>
                                                        <button
                                                            className="btn-delete"
                                                            onClick={() => handleRemoveUserFromGroup(member.id, member.username)}
                                                        >
                                                            Remove
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>
                            <div className="modal-actions">
                                <button className="btn-secondary" onClick={() => { setShowMembersModal(false); setActionError(''); }}>
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
                        onConfirm={confirm.onConfirm}
                        onCancel={() => setConfirm(null)}
                    />
                )}
            </div>
        </Layout>
    );
};

export default GroupManagement;


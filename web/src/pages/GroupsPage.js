import React, { useState, useEffect } from 'react';
import config from '../config';
import { useAuth } from '../contexts/AuthContext';
import Layout from '../components/Layout';
import './GroupsPage.css';

const GroupsPage = () => {
    const [groups, setGroups] = useState([]);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showMembersModal, setShowMembersModal] = useState(false);
    const [showAddUserModal, setShowAddUserModal] = useState(false);
    const [selectedGroup, setSelectedGroup] = useState(null);
    const [groupMembers, setGroupMembers] = useState([]);
    const { token } = useAuth();

    const [formData, setFormData] = useState({
        name: '',
        description: '',
    });

    const [selectedUserId, setSelectedUserId] = useState('');

    useEffect(() => {
        fetchGroups();
        fetchUsers();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const fetchGroups = async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await fetch(config.api.groups, {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            setGroups(data.data || []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const fetchUsers = async () => {
        try {
            const response = await fetch(config.api.users, {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            setUsers(data.users || []);
        } catch (e) {
            console.error('Failed to fetch users:', e);
        }
    };

    const fetchGroupMembers = async (groupId) => {
        try {
            const response = await fetch(config.api.groupUsers(groupId), {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            setGroupMembers(data.users || []);
        } catch (e) {
            setError(e.message);
        }
    };

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData((prev) => ({ ...prev, [name]: value }));
    };

    const resetForm = () => {
        setFormData({
            name: '',
            description: '',
        });
    };

    const handleCreateGroup = async (e) => {
        e.preventDefault();
        setError(null);

        try {
            const response = await fetch(config.api.groups, {
                method: 'POST',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to create group');
            }

            setShowCreateModal(false);
            resetForm();
            fetchGroups();
        } catch (e) {
            setError(e.message);
        }
    };

    const handleDeleteGroup = async (groupId) => {
        if (!window.confirm('Are you sure you want to delete this group?')) {
            return;
        }

        setError(null);
        try {
            const response = await fetch(config.api.groups, {
                method: 'DELETE',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ id: groupId }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to delete group');
            }

            fetchGroups();
        } catch (e) {
            setError(e.message);
        }
    };

    const handleAddUserToGroup = async (e) => {
        e.preventDefault();
        setError(null);

        if (!selectedUserId) {
            setError('Please select a user');
            return;
        }

        try {
            const response = await fetch(config.api.addUserToGroup(selectedGroup.group_id), {
                method: 'POST',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ user_id: selectedUserId }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to add user to group');
            }

            setShowAddUserModal(false);
            setSelectedUserId('');
            fetchGroupMembers(selectedGroup.group_id);
        } catch (e) {
            setError(e.message);
        }
    };

    const openMembersModal = async (group) => {
        setSelectedGroup(group);
        setShowMembersModal(true);
        await fetchGroupMembers(group.group_id);
    };

    const openAddUserModal = () => {
        setShowAddUserModal(true);
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading groups...</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="groups-page">
                <div className="page-header">
                    <h2>Group Management</h2>
                    <button className="btn-primary" onClick={() => setShowCreateModal(true)}>
                        Create Group
                    </button>
                </div>

                {error && <div className="error-message">{error}</div>}

                <div className="groups-grid">
                    {groups.map((group) => (
                        <div key={group.group_id} className="group-card">
                            <div className="group-card-header">
                                <h3>{group.group_name}</h3>
                            </div>
                            <div className="group-card-body">
                                <p className="group-id">ID: {group.group_id}</p>
                            </div>
                            <div className="group-card-actions">
                                <button
                                    className="btn-small btn-view"
                                    onClick={() => openMembersModal(group)}
                                >
                                    View Members
                                </button>
                                <button
                                    className="btn-small btn-delete"
                                    onClick={() => handleDeleteGroup(group.group_id)}
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Create Group Modal */}
                {showCreateModal && (
                    <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
                        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                            <h3>Create Group</h3>
                            <form onSubmit={handleCreateGroup}>
                                <div className="form-group">
                                    <label>Group Name:</label>
                                    <input
                                        type="text"
                                        name="name"
                                        value={formData.name}
                                        onChange={handleInputChange}
                                        required
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Description:</label>
                                    <textarea
                                        name="description"
                                        value={formData.description}
                                        onChange={handleInputChange}
                                        rows="4"
                                    />
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">
                                        Create
                                    </button>
                                    <button
                                        type="button"
                                        className="btn-secondary"
                                        onClick={() => setShowCreateModal(false)}
                                    >
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}

                {/* Group Members Modal */}
                {showMembersModal && selectedGroup && (
                    <div className="modal-overlay" onClick={() => setShowMembersModal(false)}>
                        <div
                            className="modal-content modal-large"
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h3>Members of {selectedGroup.group_name}</h3>
                            <button className="btn-primary mb-20" onClick={openAddUserModal}>
                                Add User
                            </button>

                            <div className="members-list">
                                {groupMembers.length === 0 ? (
                                    <p className="no-members">No members in this group</p>
                                ) : (
                                    <table className="members-table">
                                        <thead>
                                            <tr>
                                                <th>User ID</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {groupMembers.map((userId) => (
                                                <tr key={userId}>
                                                    <td>{userId}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>

                            <div className="modal-actions">
                                <button
                                    type="button"
                                    className="btn-secondary"
                                    onClick={() => setShowMembersModal(false)}
                                >
                                    Close
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* Add User to Group Modal */}
                {showAddUserModal && selectedGroup && (
                    <div className="modal-overlay" onClick={() => setShowAddUserModal(false)}>
                        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                            <h3>Add User to {selectedGroup.group_name}</h3>
                            <form onSubmit={handleAddUserToGroup}>
                                <div className="form-group">
                                    <label>Select User:</label>
                                    <select
                                        value={selectedUserId}
                                        onChange={(e) => setSelectedUserId(e.target.value)}
                                        required
                                    >
                                        <option value="">-- Select a user --</option>
                                        {users.map((user) => (
                                            <option key={user.id} value={user.id}>
                                                {user.username} ({user.email})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">
                                        Add User
                                    </button>
                                    <button
                                        type="button"
                                        className="btn-secondary"
                                        onClick={() => setShowAddUserModal(false)}
                                    >
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

export default GroupsPage;

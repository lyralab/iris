import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
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

    useEffect(() => {
        fetchGroups();
        fetchAllUsers();
    }, []);

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

    const fetchGroupMembers = async (groupId) => {
        try {
            const data = await apiService.getGroupUsers(groupId);
            setGroupMembers(data.users || []);
        } catch (e) {
            console.log('Error fetching group members:', e);
            setGroupMembers([]);
        }
    };

    const handleAddGroup = async (e) => {
        e.preventDefault();
        try {
            await apiService.createGroup(newGroup);
            setShowAddModal(false);
            setNewGroup({ name: '', description: '' });
            fetchGroups();
            alert('Group created successfully');
        } catch (e) {
            alert('Error creating group: ' + e.message);
        }
    };

    const handleDeleteGroup = async (group) => {
        if (!window.confirm(`Are you sure you want to delete group: ${group.group_name}?`)) return;

        try {
            await apiService.deleteGroup({ id: group.group_id });
            fetchGroups();
            alert('Group deleted successfully');
        } catch (e) {
            alert('Error deleting group: ' + e.message);
        }
    };

    const handleViewMembers = async (group) => {
        setSelectedGroup(group);
        await fetchGroupMembers(group.group_id);
        setShowMembersModal(true);
        setShowAddUserSection(false);
        setSelectedUserId('');
    };

    const handleAddUserToGroup = async () => {
        if (!selectedUserId) {
            alert('Please select a user');
            return;
        }

        try {
            await apiService.addUserToGroup(selectedGroup.group_id, selectedUserId);
            alert('User added to group successfully');
            // Refresh the group members list
            await fetchGroupMembers(selectedGroup.group_id);
            // Reset the selection
            setSelectedUserId('');
            setShowAddUserSection(false);
        } catch (e) {
            alert('Error adding user to group: ' + e.message);
        }
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
                    <button className="btn-primary" onClick={() => setShowAddModal(true)}>
                        + Add Group
                    </button>
                </div>

                <div className="groups-grid">
                    {groups.map((group) => (
                        <div key={group.group_id} className="group-card">
                            <div
                                className="group-header clickable"
                                onClick={() => handleViewMembers(group)}
                                title="Click to view members"
                            >
                                <h3>{group.group_name}</h3>
                                <span className="member-count-hint">👥 Click to view members</span>
                            </div>
                            <div className="group-actions">
                                <button
                                    className="btn-delete"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        handleDeleteGroup(group);
                                    }}
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Add Group Modal */}
                {showAddModal && (
                    <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Add New Group</h2>
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
                                    <button type="button" className="btn-secondary" onClick={() => setShowAddModal(false)}>
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
                        <div className="modal modal-large" onClick={(e) => e.stopPropagation()}>
                            <h2>Members of {selectedGroup.group_name}</h2>

                            {/* Add User Section */}
                            <div className="add-user-section">
                                {!showAddUserSection ? (
                                    <button
                                        className="btn-add-user"
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
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {groupMembers.map((member) => (
                                                <tr key={member.id || member.username}>
                                                    <td>{member.username}</td>
                                                    <td>{member.email || '-'}</td>
                                                    <td>{member.mobile || '-'}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>
                            <div className="modal-actions">
                                <button className="btn-secondary" onClick={() => setShowMembersModal(false)}>
                                    Close
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </Layout>
    );
};

export default GroupManagement;


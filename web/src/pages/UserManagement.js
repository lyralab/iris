import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
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
        password: '',
        'confirm-password': ''
    });

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
        
        if (newUser.password !== newUser['confirm-password']) {
            alert('Passwords do not match');
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
                password: '',
                'confirm-password': ''
            });
            fetchUsers();
            alert('User added successfully');
        } catch (e) {
            alert('Error adding user: ' + e.message);
        }
    };

    const handleUpdateUser = async (e) => {
        e.preventDefault();
        try {
            await apiService.updateUser(selectedUser);
            setShowEditModal(false);
            setSelectedUser(null);
            fetchUsers();
            alert('User updated successfully');
        } catch (e) {
            alert('Error updating user: ' + e.message);
        }
    };

    const handleVerifyUser = async (username) => {
        if (!window.confirm(`Are you sure you want to verify user: ${username}?`)) return;

        try {
            await apiService.verifyUser(username);
            fetchUsers();
            alert('User verified successfully');
        } catch (e) {
            alert('Error verifying user: ' + e.message);
        }
    };

    const handleAddUserToGroup = async (groupId) => {
        try {
            await apiService.addUserToGroup(groupId, selectedUserForGroup.id);
            alert('User added to group successfully');
            setShowGroupModal(false);
            setSelectedUserForGroup(null);
        } catch (e) {
            alert('Error adding user to group: ' + e.message);
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

    return (
        <Layout>
            <div className="user-management">
                <div className="page-header">
                    <h1>User Management</h1>
                    <button className="btn-primary" onClick={() => setShowAddModal(true)}>
                        + Add User
                    </button>
                </div>

                <div className="users-table-container">
                    <table className="users-table">
                        <thead>
                            <tr>
                                <th>Username</th>
                                <th>First Name</th>
                                <th>Last Name</th>
                                <th>Email</th>
                                <th>Mobile</th>
                                <th>Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.map((user) => (
                                <tr key={user.id}>
                                    <td>{user.username}</td>
                                    <td>{user.firstName || '-'}</td>
                                    <td>{user.lastName || '-'}</td>
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
                                                setSelectedUser(user);
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

                {/* Add User Modal */}
                {showAddModal && (
                    <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
                        <div className="modal" onClick={(e) => e.stopPropagation()}>
                            <h2>Add New User</h2>
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
                                            value={selectedUser.firstName || ''}
                                            onChange={(e) => setSelectedUser({ ...selectedUser, firstName: e.target.value })}
                                        />
                                    </div>
                                    <div className="form-group">
                                        <label>Last Name:</label>
                                        <input
                                            type="text"
                                            value={selectedUser.lastName || ''}
                                            onChange={(e) => setSelectedUser({ ...selectedUser, lastName: e.target.value })}
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
                                        value={selectedUser.mobile || ''}
                                        onChange={(e) => setSelectedUser({ ...selectedUser, mobile: e.target.value })}
                                        pattern="[0-9]{11}"
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
                            <div className="groups-list">
                                {groups.map((group) => (
                                    <div key={group.GroupID} className="group-item">
                                        <span>{group.GroupName}</span>
                                        <button
                                            className="btn-primary"
                                            onClick={() => handleAddUserToGroup(group.GroupID)}
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
            </div>
        </Layout>
    );
};

export default UserManagement;


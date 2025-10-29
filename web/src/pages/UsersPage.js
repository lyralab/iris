import React, { useState, useEffect } from 'react';
import config from '../config';
import { useAuth } from '../contexts/AuthContext';
import Layout from '../components/Layout';
import './UsersPage.css';

const UsersPage = () => {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showModal, setShowModal] = useState(false);
    const [editingUser, setEditingUser] = useState(null);
    const { token } = useAuth();

    const [formData, setFormData] = useState({
        username: '',
        firstname: '',
        lastname: '',
        email: '',
        password: '',
        'confirm-password': '',
    });

    useEffect(() => {
        fetchUsers();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const fetchUsers = async () => {
        setLoading(true);
        setError(null);
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
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData((prev) => ({ ...prev, [name]: value }));
    };

    const resetForm = () => {
        setFormData({
            username: '',
            firstname: '',
            lastname: '',
            email: '',
            password: '',
            'confirm-password': '',
        });
        setEditingUser(null);
    };

    const handleCreateUser = async (e) => {
        e.preventDefault();
        setError(null);

        if (formData.password !== formData['confirm-password']) {
            setError('Passwords do not match');
            return;
        }

        try {
            const response = await fetch(config.api.users, {
                method: 'POST',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to create user');
            }

            setShowModal(false);
            resetForm();
            fetchUsers();
        } catch (e) {
            setError(e.message);
        }
    };

    const handleUpdateUser = async (e) => {
        e.preventDefault();
        setError(null);

        const updateData = {
            username: formData.username,
            firstname: formData.firstname,
            lastname: formData.lastname,
            email: formData.email,
        };

        if (formData.password) {
            updateData.password = formData.password;
        }

        try {
            const response = await fetch(config.api.users, {
                method: 'PUT',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updateData),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to update user');
            }

            setShowModal(false);
            resetForm();
            fetchUsers();
        } catch (e) {
            setError(e.message);
        }
    };

    const handleVerifyUser = async (username) => {
        setError(null);
        try {
            const response = await fetch(config.api.verifyUser, {
                method: 'PUT',
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to verify user');
            }

            fetchUsers();
        } catch (e) {
            setError(e.message);
        }
    };

    const openCreateModal = () => {
        resetForm();
        setShowModal(true);
    };

    const openEditModal = (user) => {
        setEditingUser(user);
        setFormData({
            username: user.username,
            firstname: user.firstName || '',
            lastname: user.lastName || '',
            email: user.email || '',
            password: '',
            'confirm-password': '',
        });
        setShowModal(true);
    };

    if (loading) {
        return (
            <Layout>
                <div className="loading">Loading users...</div>
            </Layout>
        );
    }

    return (
        <Layout>
            <div className="users-page">
                <div className="page-header">
                    <h2>User Management</h2>
                    <button className="btn-primary" onClick={openCreateModal}>
                        Create User
                    </button>
                </div>

                {error && <div className="error-message">{error}</div>}

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
                                        <span className={`status-badge ${user.status}`}>
                                            {user.status}
                                        </span>
                                    </td>
                                    <td className="actions">
                                        <button
                                            className="btn-small btn-edit"
                                            onClick={() => openEditModal(user)}
                                        >
                                            Edit
                                        </button>
                                        {user.status === 'pending' && (
                                            <button
                                                className="btn-small btn-verify"
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

                {showModal && (
                    <div className="modal-overlay" onClick={() => setShowModal(false)}>
                        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                            <h3>{editingUser ? 'Edit User' : 'Create User'}</h3>
                            <form onSubmit={editingUser ? handleUpdateUser : handleCreateUser}>
                                <div className="form-group">
                                    <label>Username:</label>
                                    <input
                                        type="text"
                                        name="username"
                                        value={formData.username}
                                        onChange={handleInputChange}
                                        required
                                        disabled={editingUser !== null}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>First Name:</label>
                                    <input
                                        type="text"
                                        name="firstname"
                                        value={formData.firstname}
                                        onChange={handleInputChange}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Last Name:</label>
                                    <input
                                        type="text"
                                        name="lastname"
                                        value={formData.lastname}
                                        onChange={handleInputChange}
                                    />
                                </div>
                                <div className="form-group">
                                    <label>Email:</label>
                                    <input
                                        type="email"
                                        name="email"
                                        value={formData.email}
                                        onChange={handleInputChange}
                                    />
                                </div>
                                {!editingUser && (
                                    <>
                                        <div className="form-group">
                                            <label>Password:</label>
                                            <input
                                                type="password"
                                                name="password"
                                                value={formData.password}
                                                onChange={handleInputChange}
                                                required
                                            />
                                        </div>
                                        <div className="form-group">
                                            <label>Confirm Password:</label>
                                            <input
                                                type="password"
                                                name="confirm-password"
                                                value={formData['confirm-password']}
                                                onChange={handleInputChange}
                                                required
                                            />
                                        </div>
                                    </>
                                )}
                                {editingUser && (
                                    <div className="form-group">
                                        <label>New Password (optional):</label>
                                        <input
                                            type="password"
                                            name="password"
                                            value={formData.password}
                                            onChange={handleInputChange}
                                        />
                                    </div>
                                )}
                                <div className="modal-actions">
                                    <button type="submit" className="btn-primary">
                                        {editingUser ? 'Update' : 'Create'}
                                    </button>
                                    <button
                                        type="button"
                                        className="btn-secondary"
                                        onClick={() => setShowModal(false)}
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

export default UsersPage;

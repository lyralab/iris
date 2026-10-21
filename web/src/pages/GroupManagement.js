import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import ConfirmDialog from '../components/ConfirmDialog';
import EmptyState from '../components/EmptyState';
import Icon from '../components/Icon';
import Modal from '../components/Modal';
import { initials } from '../utils/format';
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

    const closeMembers = () => { setShowMembersModal(false); setActionError(''); };
    const closeAdd = () => { setShowAddModal(false); setActionError(''); };
    const candidates = allUsers.filter(user => !groupMembers.some(member => member.id === user.id));

    let content;
    if (loading) {
        content = <div className="loading">Loading groups…</div>;
    } else if (error) {
        content = <div className="error-message"><Icon name="alert" />Error loading groups: {error}</div>;
    } else if (groups.length === 0) {
        content = (
            <div className="card">
                <EmptyState icon="group" title="No groups yet">Groups route alerts to a set of people at once.</EmptyState>
            </div>
        );
    } else {
        content = (
            <div className="groups-grid">
                {groups.map((group) => (
                    <article key={group.group_id} className="card group-card">
                        <div className="group-body">
                            <span className="group-icon"><Icon name="group" size={18} /></span>
                            <div className="group-text">
                                <h3>{group.group_name}</h3>
                                <p className="group-description">{group.description || 'No description'}</p>
                            </div>
                        </div>
                        <div className="group-actions">
                            <button className="btn-secondary btn-sm" onClick={() => handleViewMembers(group)}>
                                <Icon name="users" size={14} />
                                Members
                            </button>
                            <button
                                className="btn-danger-soft btn-sm btn-icon"
                                onClick={() => handleDeleteGroup(group)}
                                aria-label={`Delete ${group.group_name}`}
                                title="Delete group"
                            >
                                <Icon name="trash" size={15} />
                            </button>
                        </div>
                    </article>
                ))}
            </div>
        );
    }

    return (
        <Layout>
            <div className="page group-management">
                <div className="page-header">
                    <div>
                        <h1>Groups</h1>
                        <p className="subtitle">Teams that receive alerts together</p>
                    </div>
                    <button className="btn-primary" onClick={() => { setActionError(''); setShowAddModal(true); }}>
                        <Icon name="plus" />
                        New group
                    </button>
                </div>

                {actionError && !showAddModal && !showMembersModal && (
                    <div className="error-message"><Icon name="alert" />{actionError}</div>
                )}

                {content}

                {showAddModal && (
                    <Modal title="New group" onClose={closeAdd}>
                        <form onSubmit={handleAddGroup}>
                            <div className="modal-body">
                                {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}
                                <div className="form-group">
                                    <label htmlFor="group-name">Group name</label>
                                    <input
                                        id="group-name"
                                        type="text"
                                        value={newGroup.name}
                                        onChange={(e) => setNewGroup({ ...newGroup, name: e.target.value })}
                                        required
                                        minLength="3"
                                        placeholder="e.g. payments-oncall"
                                    />
                                </div>
                                <div className="form-group">
                                    <label htmlFor="group-description">Description</label>
                                    <textarea
                                        id="group-description"
                                        value={newGroup.description}
                                        onChange={(e) => setNewGroup({ ...newGroup, description: e.target.value })}
                                        rows="3"
                                        placeholder="What this group is responsible for"
                                    />
                                </div>
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn-secondary" onClick={closeAdd}>Cancel</button>
                                <button type="submit" className="btn-primary">Create group</button>
                            </div>
                        </form>
                    </Modal>
                )}

                {showMembersModal && selectedGroup && (
                    <Modal
                        title={selectedGroup.group_name}
                        subtitle={`${groupMembers.length} ${groupMembers.length === 1 ? 'member' : 'members'}`}
                        size="modal-large"
                        onClose={closeMembers}
                    >
                        <div className="modal-body">
                            {actionError && <div className="error-message"><Icon name="alert" />{actionError}</div>}

                            <div className="add-user-section">
                                {!showAddUserSection ? (
                                    <button className="btn-secondary btn-add-user" onClick={() => setShowAddUserSection(true)}>
                                        <Icon name="userPlus" />
                                        Add member
                                    </button>
                                ) : (
                                    <div className="add-user-form">
                                        <select
                                            value={selectedUserId}
                                            onChange={(e) => setSelectedUserId(e.target.value)}
                                            className="select-input user-dropdown"
                                            aria-label="User to add"
                                            autoFocus
                                        >
                                            <option value="">Select a user…</option>
                                            {candidates.map(user => (
                                                <option key={user.id} value={user.id}>
                                                    {user.username} ({user.email || user.mobile || 'No contact'})
                                                </option>
                                            ))}
                                        </select>
                                        <button className="btn-primary" onClick={handleAddUserToGroup} disabled={!selectedUserId}>
                                            Add
                                        </button>
                                        <button
                                            className="btn-ghost"
                                            onClick={() => {
                                                setShowAddUserSection(false);
                                                setSelectedUserId('');
                                            }}
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                )}
                            </div>

                            <div className="members-list">
                                {groupMembers.length === 0 ? (
                                    <EmptyState icon="users" title="No members yet">Add people so this group receives alerts.</EmptyState>
                                ) : (
                                    <table className="members-table">
                                        <thead>
                                            <tr>
                                                <th>Member</th>
                                                <th>Email</th>
                                                <th>Mobile</th>
                                                <th className="cell-actions"><span className="visually-hidden">Actions</span></th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {groupMembers.map((member) => (
                                                <tr key={member.id || member.username}>
                                                    <td>
                                                        <span className="member-cell">
                                                            <span className="avatar">{initials(member.username)}</span>
                                                            {member.username}
                                                        </span>
                                                    </td>
                                                    <td className="cell-muted">{member.email || '—'}</td>
                                                    <td className="cell-muted tabular">{member.mobile || '—'}</td>
                                                    <td className="cell-actions">
                                                        <button
                                                            className="btn-danger-soft btn-sm"
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
                        </div>
                        <div className="modal-footer">
                            <button className="btn-secondary" onClick={closeMembers}>Close</button>
                        </div>
                    </Modal>
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

import React, { useState, useEffect } from 'react';
import apiService from '../utils/apiService';
import Layout from '../components/Layout';
import Icon from '../components/Icon';
import Modal from '../components/Modal';
import { initials, userStatusInfo } from '../utils/format';
import './Profile.css';

// Shows `value` (wrapped in `className`), or children for custom content; "Not set" when empty.
const Field = ({ label, value, className, children }) => {
    const content = children ?? (value ? <span className={className}>{value}</span> : null);
    return (
        <div className="info-row">
            <dt>{label}</dt>
            <dd>{content || <span className="cell-muted">Not set</span>}</dd>
        </div>
    );
};

const EDIT_FIELDS = [
    { name: 'firstname', label: 'First name', type: 'text' },
    { name: 'lastname', label: 'Last name', type: 'text' },
    { name: 'email', label: 'Email', type: 'email' },
    { name: 'mobile_number', label: 'Mobile', type: 'text', placeholder: '09xxxxxxxxx' },
    { name: 'telegram_id', label: 'Telegram ID', type: 'text', pattern: '-?[0-9]+', maxLength: 50, placeholder: 'User or channel ID' },
    { name: 'mattermost_id', label: 'Mattermost ID', type: 'text', maxLength: 50, placeholder: 'Mattermost user or channel ID' },
    { name: 'bale_id', label: 'Bale ID', type: 'text', pattern: '-?[0-9]+', maxLength: 50, placeholder: 'User or channel ID' },
];

const Profile = () => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showEditModal, setShowEditModal] = useState(false);
    const [editForm, setEditForm] = useState({
        username: '',
        firstname: '',
        lastname: '',
        email: '',
        mobile_number: '',
        telegram_id: '',
        mattermost_id: '',
        bale_id: ''
    });
    const [saving, setSaving] = useState(false);
    const [saveError, setSaveError] = useState(null);

    const fetchUserProfile = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiService.getUserMe();
            setUser(data.user || data);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUserProfile();
    }, []);

    const openEditModal = () => {
        setEditForm({
            username: user.username || '',
            firstname: user.firstName || user.firstname || '',
            lastname: user.lastName || user.lastname || '',
            email: user.email || '',
            mobile_number: user.mobile || user.mobile_number || '',
            telegram_id: user.telegramID || user.telegram_id || '',
            mattermost_id: user.mattermostID || user.mattermost_id || '',
            bale_id: user.baleID || user.bale_id || ''
        });
        setSaveError(null);
        setShowEditModal(true);
    };

    const closeEditModal = () => {
        setShowEditModal(false);
        setSaveError(null);
    };

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setEditForm(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleSave = async () => {
        setSaving(true);
        setSaveError(null);
        try {
            await apiService.updateUser(editForm);
            await fetchUserProfile();
            setShowEditModal(false);
        } catch (e) {
            setSaveError(e.message);
        } finally {
            setSaving(false);
        }
    };

    let body;
    if (loading) {
        body = <div className="loading">Loading profile…</div>;
    } else if (error) {
        body = <div className="error-message"><Icon name="alert" />Error loading profile: {error}</div>;
    } else if (!user) {
        body = <div className="error-message"><Icon name="alert" />User information not available</div>;
    } else {
        const statusInfo = userStatusInfo(user.status);
        const first = user.firstName || user.firstname;
        const last = user.lastName || user.lastname;
        const fullName = [first, last].filter(Boolean).join(' ');
        body = (
            <>
                <section className="card profile-hero">
                    <span className="avatar avatar-lg">{fullName ? initials(first, last) : initials(user.username)}</span>
                    <div className="profile-hero-text">
                        <h1>{fullName || user.username}</h1>
                        <div className="profile-hero-meta">
                            <span className="cell-muted">@{user.username}</span>
                            <span className="badge badge-outline">{user.role || 'member'}</span>
                            <span className={`status-badge ${statusInfo.className}`}>{statusInfo.label}</span>
                        </div>
                    </div>
                    <button className="btn-secondary" onClick={openEditModal}>
                        <Icon name="edit" />
                        Edit profile
                    </button>
                </section>

                <div className="profile-grid">
                    <section className="card">
                        <div className="card-header"><h2 className="card-title"><Icon name="user" />Personal</h2></div>
                        <dl className="info-list card-body">
                            <Field label="Username" value={user.username} />
                            <Field label="First name" value={first} />
                            <Field label="Last name" value={last} />
                            <Field label="Role" value={user.role} className="capitalize" />
                        </dl>
                    </section>

                    <section className="card">
                        <div className="card-header"><h2 className="card-title"><Icon name="bell" />Notification channels</h2></div>
                        <dl className="info-list card-body">
                            <Field label="Email" value={user.email} />
                            <Field label="Mobile" value={user.mobile || user.mobile_number} className="tabular" />
                            <Field label="Telegram ID" value={user.telegramID || user.telegram_id} className="mono" />
                            <Field label="Mattermost ID" value={user.mattermostID || user.mattermost_id} className="mono" />
                            <Field label="Bale ID" value={user.baleID || user.bale_id} className="mono" />
                        </dl>
                    </section>

                    <section className="card">
                        <div className="card-header"><h2 className="card-title"><Icon name="shield" />Account</h2></div>
                        <dl className="info-list card-body">
                            <Field label="Status"><span className={`status-badge ${statusInfo.className}`}>{statusInfo.label}</span></Field>
                            <Field
                                label="Member since"
                                value={user.createdAt ? new Date(user.createdAt).toLocaleDateString('en-US', {
                                    year: 'numeric',
                                    month: 'long',
                                    day: 'numeric'
                                }) : null}
                            />
                        </dl>
                    </section>
                </div>
            </>
        );
    }

    return (
        <Layout>
            <div className="page profile-page">
                {body}

                {showEditModal && (
                    <Modal title="Edit profile" onClose={closeEditModal}>
                        <form onSubmit={(e) => { e.preventDefault(); handleSave(); }}>
                            <div className="modal-body">
                                {saveError && <div className="modal-error"><Icon name="alert" />{saveError}</div>}
                                <div className="form-row">
                                    {EDIT_FIELDS.map((f) => (
                                        <div className="form-group" key={f.name}>
                                            <label htmlFor={`profile-${f.name}`}>{f.label}</label>
                                            <input
                                                id={`profile-${f.name}`}
                                                type={f.type}
                                                name={f.name}
                                                value={editForm[f.name]}
                                                onChange={handleInputChange}
                                                pattern={f.pattern}
                                                maxLength={f.maxLength}
                                                placeholder={f.placeholder}
                                            />
                                        </div>
                                    ))}
                                </div>
                            </div>
                            <div className="modal-footer">
                                <button type="button" className="btn-secondary" onClick={closeEditModal}>Cancel</button>
                                <button type="submit" className="btn-primary" disabled={saving}>
                                    {saving ? 'Saving…' : 'Save changes'}
                                </button>
                            </div>
                        </form>
                    </Modal>
                )}
            </div>
        </Layout>
    );
};

export default Profile;

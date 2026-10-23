import React, { useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { isAuthenticated, mustChangePassword, setCookie, logout } from '../utils/auth';
import apiService from '../utils/apiService';
import Icon from '../components/Icon';
import Logo from '../components/Logo';
import ThemeToggle from '../components/ThemeToggle';
import '../LoginPage.css';

const FIELDS = [
    { name: 'current', label: 'Current password', autoComplete: 'current-password' },
    { name: 'next', label: 'New password', autoComplete: 'new-password' },
    { name: 'confirm', label: 'Confirm new password', autoComplete: 'new-password' },
];

// Standalone (no Layout) so a user who must change their password sees nothing else.
const ChangePassword = () => {
    const [form, setForm] = useState({ current: '', next: '', confirm: '' });
    const [error, setError] = useState('');
    const [saving, setSaving] = useState(false);
    const navigate = useNavigate();

    if (!isAuthenticated()) {
        return <Navigate to="/" replace />;
    }
    const forced = mustChangePassword();

    const handleSubmit = async (event) => {
        event.preventDefault();
        setError('');
        if (form.next !== form.confirm) {
            setError('New passwords do not match.');
            return;
        }
        setSaving(true);
        try {
            const data = await apiService.changePassword(form.current, form.next, form.confirm);
            setCookie('jwt', data.token, 1);
            navigate('/dashboard', { replace: true });
        } catch (err) {
            setError(err.message || 'Password change failed');
        } finally {
            setSaving(false);
        }
    };

    const handleLeave = () => {
        if (forced) {
            logout();
            navigate('/', { replace: true });
        } else {
            navigate('/profile');
        }
    };

    return (
        <div className="login-page">
            <main className="login-panel">
                <div className="login-theme"><ThemeToggle /></div>
                <div className="login-container">
                    <div className="login-header">
                        <Logo tile size={44} title="Iris" className="login-logo" />
                        <h2>Change password</h2>
                        <p className="login-subtitle">
                            {forced
                                ? 'Your password was set by someone else. Choose a new one to continue.'
                                : 'Choose a new password for your account.'}
                        </p>
                    </div>

                    {error && (
                        <div className="error" role="alert">
                            <Icon name="alert" />
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleSubmit}>
                        {FIELDS.map((f) => (
                            <div className="form-group" key={f.name}>
                                <label htmlFor={`pw-${f.name}`}>{f.label}</label>
                                <div className="input-with-icon">
                                    <Icon name="lock" />
                                    <input
                                        type="password"
                                        id={`pw-${f.name}`}
                                        value={form[f.name]}
                                        onChange={(e) => setForm((prev) => ({ ...prev, [f.name]: e.target.value }))}
                                        autoComplete={f.autoComplete}
                                        minLength={f.name === 'current' ? undefined : 8}
                                        maxLength={30}
                                        required
                                        disabled={saving}
                                    />
                                </div>
                            </div>
                        ))}
                        <p className="login-subtitle">
                            8–30 characters with upper and lower case letters, a digit and one of @$!%*?&#^
                        </p>

                        <button type="submit" className="btn-primary login-button" disabled={saving}>
                            {saving ? 'Saving…' : 'Change password'}
                        </button>
                        <button type="button" className="btn-ghost login-button" onClick={handleLeave} disabled={saving}>
                            {forced ? 'Sign out' : 'Cancel'}
                        </button>
                    </form>
                </div>
            </main>
        </div>
    );
};

export default ChangePassword;

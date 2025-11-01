import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { setCookie, isAuthenticated } from './utils/auth';
import apiService from './utils/apiService';
import './LoginPage.css';


const LoginPage = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [captchaAnswer, setCaptchaAnswer] = useState('');
    const [captchaId, setCaptchaId] = useState('');
    const [captchaImage, setCaptchaImage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const [captchaLoading, setCaptchaLoading] = useState(false);
    const navigate = useNavigate();

    useEffect(() => {
        if (isAuthenticated()) {
            navigate('/dashboard');
        } else {
            fetchCaptcha();
        }
    }, [navigate]);

    const fetchCaptcha = async () => {
        setCaptchaLoading(true);
        setError('');
        try {
            const data = await apiService.generateCaptcha();
            if (data.status === 'success' && data.data) {
                setCaptchaId(data.data.id);
                setCaptchaImage(data.data.b64);
            } else {
                setError('Failed to load captcha');
            }
        } catch (err) {
            setError('Network error loading captcha');
        } finally {
            setCaptchaLoading(false);
        }
    };

    const handleRefreshCaptcha = () => {
        setCaptchaAnswer('');
        fetchCaptcha();
    };

    const handleLogin = async (event) => {
        event.preventDefault();
        setError('');

        if (!captchaAnswer) {
            setError('Please enter the captcha answer.');
            return;
        }

        setLoading(true);

        try {
            const data = await apiService.signin(username, password, captchaId, captchaAnswer);
            if (data.token) {
                setCookie('jwt', data.token, 1);
                navigate('/dashboard');
            } else {
                setError('Login failed: Token not received.');
                handleRefreshCaptcha();
            }
        } catch (err) {
            setError(err.message || 'Login failed');
            handleRefreshCaptcha();
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <div className="login-container">
                <div className="login-header">
                    <img src="/iris.png" alt="Iris Logo" className="login-logo" />
                    <h2>Iris Alert Manager</h2>
                    <p className="login-subtitle">Sign in to continue</p>
                </div>

                {error && <div className="error">{error}</div>}

                <form onSubmit={handleLogin}>
                    <div className="form-group">
                        <label htmlFor="username">
                            <span className="label-icon">👤</span>
                            Username
                        </label>
                        <input
                            type="text"
                            id="username"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="Enter your username"
                            required
                            disabled={loading}
                        />
                    </div>

                    <div className="form-group">
                        <label htmlFor="password">
                            <span className="label-icon">🔒</span>
                            Password
                        </label>
                        <input
                            type="password"
                            id="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Enter your password"
                            required
                            disabled={loading}
                        />
                    </div>

                    <div className="form-group captcha-group">
                        <label htmlFor="captcha">
                            <span className="label-icon">🔐</span>
                            Security Verification
                        </label>
                        <div className="captcha-container">
                            <div className="captcha-image-wrapper">
                                {captchaLoading ? (
                                    <div className="captcha-loading">Loading...</div>
                                ) : captchaImage ? (
                                    <img
                                        src={captchaImage}
                                        alt="Captcha"
                                        className="captcha-image"
                                    />
                                ) : (
                                    <div className="captcha-error">Failed to load</div>
                                )}
                            </div>
                            <button
                                type="button"
                                className="captcha-refresh"
                                onClick={handleRefreshCaptcha}
                                disabled={captchaLoading || loading}
                                title="Refresh captcha"
                            >
                                🔄
                            </button>
                        </div>
                        <input
                            type="text"
                            id="captcha"
                            value={captchaAnswer}
                            onChange={(e) => setCaptchaAnswer(e.target.value)}
                            placeholder="Enter the answer"
                            required
                            disabled={loading || !captchaId}
                            autoComplete="off"
                        />
                        <small className="captcha-hint">Solve the math problem shown in the image</small>
                    </div>

                    <button
                        type="submit"
                        className="login-button"
                        disabled={loading || captchaLoading || !captchaId}
                    >
                        {loading ? 'Signing in...' : 'Sign In'}
                    </button>
                </form>

                <div className="login-footer">
                    <p>© 2025 Iris Alert Manager. All rights reserved.</p>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { setCookie, isAuthenticated } from './utils/auth';
import apiService from './utils/apiService';
import Icon from './components/Icon';
import Logo from './components/Logo';
import ThemeToggle from './components/ThemeToggle';
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

    const fetchCaptcha = async (clearError = true) => {
        setCaptchaLoading(true);
        if (clearError) {
            setError('');
        }
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

    useEffect(() => {
        if (isAuthenticated()) {
            navigate('/dashboard');
        } else {
            fetchCaptcha();
        }
    }, [navigate]);

    const handleRefreshCaptcha = () => {
        setCaptchaAnswer('');
        fetchCaptcha(false); // Don't clear error when manually refreshing
    };

    const handleLogin = async (event) => {
        event.preventDefault();
        event.stopPropagation();
        
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
            const errorMessage = err.message || 'Login failed';
            setError(errorMessage);
            setCaptchaAnswer('');
            
            // Refresh captcha but keep the error message visible
            setTimeout(() => {
                fetchCaptcha(false);
            }, 100);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-page">
            <main className="login-panel">
                <div className="login-theme"><ThemeToggle /></div>
                <div className="login-container">
                    <div className="login-header">
                        <Logo tile size={44} title="Iris" className="login-logo" />
                        <h2>Sign in to Iris</h2>
                        <p className="login-subtitle">Alert management console</p>
                    </div>

                    {error && (
                        <div className="error" role="alert">
                            <Icon name="alert" />
                            {error}
                        </div>
                    )}

                    <form onSubmit={handleLogin}>
                        <div className="form-group">
                            <label htmlFor="username">Username</label>
                            <div className="input-with-icon">
                                <Icon name="user" />
                                <input
                                    type="text"
                                    id="username"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder="Enter your username"
                                    autoComplete="username"
                                    required
                                    disabled={loading}
                                />
                            </div>
                        </div>

                        <div className="form-group">
                            <label htmlFor="password">Password</label>
                            <div className="input-with-icon">
                                <Icon name="lock" />
                                <input
                                    type="password"
                                    id="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder="Enter your password"
                                    autoComplete="current-password"
                                    required
                                    disabled={loading}
                                />
                            </div>
                        </div>

                        <div className="form-group captcha-group">
                            <label htmlFor="captcha">Security check</label>
                            <div className="captcha-container">
                                <div className="captcha-image-wrapper">
                                    {captchaLoading ? (
                                        <span className="skeleton captcha-loading" />
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
                                    title="New captcha"
                                    aria-label="Load a new captcha"
                                >
                                    <Icon name="refresh" size={18} className={captchaLoading ? 'spin' : ''} />
                                </button>
                            </div>
                            <input
                                type="text"
                                id="captcha"
                                value={captchaAnswer}
                                onChange={(e) => setCaptchaAnswer(e.target.value)}
                                placeholder="Answer to the math problem"
                                required
                                disabled={loading || !captchaId}
                                autoComplete="off"
                                inputMode="numeric"
                            />
                        </div>

                        <button
                            type="submit"
                            className="btn-primary login-button"
                            disabled={loading || captchaLoading || !captchaId}
                        >
                            {loading ? 'Signing in…' : 'Sign in'}
                        </button>
                    </form>

                    <p className="login-footer">© {new Date().getFullYear()} Iris Alert Manager</p>
                </div>
            </main>
        </div>
    );
};

export default LoginPage;

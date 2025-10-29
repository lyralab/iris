import React, { useState, useEffect } from 'react';
import config from './config';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './contexts/AuthContext';
import './LoginPage.css';

// Simple JWT decode function (only decode payload, don't verify)
const decodeJWT = (token) => {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            throw new Error('Invalid JWT format');
        }
        const payload = parts[1];
        const decoded = JSON.parse(atob(payload));
        return decoded;
    } catch (error) {
        console.error('Failed to decode JWT:', error);
        return null;
    }
};

const LoginPage = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [captchaAnswer, setCaptchaAnswer] = useState('');
    const [captchaId, setCaptchaId] = useState('');
    const [captchaImage, setCaptchaImage] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();
    const { login } = useAuth();

    useEffect(() => {
        fetchCaptcha();
    }, []);

    const fetchCaptcha = async () => {
        try {
            const response = await fetch(`${config.api.captcha}/generate`, {
                method: 'GET',
            });

            if (response.ok) {
                const data = await response.json();
                setCaptchaId(data.captcha_id);
                setCaptchaImage(data.captcha_image);
            } else {
                console.error('Failed to fetch captcha');
            }
        } catch (err) {
            console.error('Error fetching captcha:', err);
        }
    };

    const handleLogin = async (event) => {
        event.preventDefault();
        setError('');
        setLoading(true);

        try {
            const response = await fetch(`${config.api.signin}?captcha_id=${captchaId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    username, 
                    password,
                    captcha_answer: captchaAnswer 
                }),
            });

            if (response.ok) {
                const data = await response.json();
                if (data.token) {
                    // Decode JWT to get user role
                    const decodedToken = decodeJWT(data.token);
                    
                    if (!decodedToken || !decodedToken.role) {
                        setError('Invalid token received.');
                        setLoading(false);
                        return;
                    }

                    const userData = {
                        username: decodedToken.username || username,
                        role: decodedToken.role,
                    };

                    // Only allow admin users to login
                    if (userData.role !== 'admin') {
                        setError('Access denied. Only admin users can login.');
                        fetchCaptcha(); // Refresh captcha
                        setLoading(false);
                        return;
                    }

                    login(data.token, userData);
                    navigate('/alerts');
                } else {
                    setError('Login failed: Token not received.');
                }
            } else {
                const errorData = await response.json();
                setError(errorData.message || 'Login failed!');
                fetchCaptcha(); // Refresh captcha on error
            }
        } catch (err) {
            setError('Network error. Please try again.');
            fetchCaptcha(); // Refresh captcha on error
        } finally {
            setLoading(false);
            setCaptchaAnswer(''); // Clear captcha answer
        }
    };

    return (
        <div className="login-page">
            <h2>Admin Login</h2>
            {error && <div className="error">{error}</div>}
            <form onSubmit={handleLogin}>
                <div className="form-group">
                    <label htmlFor="username">Username:</label>
                    <input
                        type="text"
                        id="username"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        disabled={loading}
                    />
                </div>
                <div className="form-group">
                    <label htmlFor="password">Password:</label>
                    <input
                        type="password"
                        id="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        disabled={loading}
                    />
                </div>
                {captchaImage && (
                    <div className="form-group">
                        <label>Captcha:</label>
                        <img 
                            src={`data:image/png;base64,${captchaImage}`} 
                            alt="Captcha" 
                            className="captcha-image"
                        />
                        <input
                            type="text"
                            placeholder="Enter captcha"
                            value={captchaAnswer}
                            onChange={(e) => setCaptchaAnswer(e.target.value)}
                            required
                            disabled={loading}
                        />
                    </div>
                )}
                <button type="submit" disabled={loading}>
                    {loading ? 'Logging in...' : 'Login'}
                </button>
            </form>
        </div>
    );
};

export default LoginPage;
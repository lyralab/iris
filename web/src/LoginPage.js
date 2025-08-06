import React, { useState, useEffect } from 'react';
import config from './config';
import { useNavigate } from 'react-router-dom';
import './LoginPage.css';


const LoginPage = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleLogin = async (event) => {
        event.preventDefault();
        setError('');

        try {
            const response = await fetch(config.api.signin, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            if (response.ok) {
                // Check if the response contains a token
                const data = await response.json();
                if (data.token) {
                    // Store the token in a cookie
                    document.cookie = `jwt=${data.token}; path=/; max-age=3600; HttpOnly`;
                    // Redirect to the alerts dashboard
                    navigate('/alerts');
                } else {
                    setError('Login failed: Token not received.');
                }
            } else {
                const errorData = await response.json();
                setError(errorData.message || 'Login failed!');
            }
        } catch (err) {
            setError('Network error. Please try again.');
        }
    };

    return (
        <div className="login-page">
            <h2>Login</h2>
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
                    />
                </div>
                <button type="submit">Login</button>
            </form>
        </div>
    );
};

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

export default LoginPage;
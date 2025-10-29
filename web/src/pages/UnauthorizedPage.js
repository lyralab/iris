import React from 'react';
import { Link } from 'react-router-dom';
import './UnauthorizedPage.css';

const UnauthorizedPage = () => {
    return (
        <div className="unauthorized-page">
            <div className="unauthorized-content">
                <h1>403</h1>
                <h2>Unauthorized Access</h2>
                <p>You do not have permission to access this page.</p>
                <p>Only admin users can access this application.</p>
                <Link to="/" className="back-link">
                    Go back to login
                </Link>
            </div>
        </div>
    );
};

export default UnauthorizedPage;

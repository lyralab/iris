import React, { useState, useEffect } from 'react';
import config from './config';
import { useAuth } from './contexts/AuthContext';

const FiringIssues = () => {
    const [severityFilter, setSeverityFilter] = useState('critical');
    const [issues, setIssues] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { token } = useAuth();

    useEffect(() => {
        const fetchFiringIssues = async () => {
            setLoading(true);
            setError(null);

            try {
                const response = await fetch(config.api.firingIssues, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                });
                
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const data = await response.json();

                const issueCounts = data.alerts.reduce((acc, alert) => {
                    const name = alert.name;
                    if (!acc[name] || new Date(alert.starts_at) < new Date(acc[name].starts_at)) {
                        acc[name] = {
                            name: name,
                            severity: alert.severity,
                            starts_at: alert.starts_at,
                            description: alert.description,
                        };
                    }
                    return acc;
                }, {});

                const sortedIssues = Object.values(issueCounts)
                    .sort((a, b) => b.count - a.count)
                    .slice(0, 10);
                setIssues(sortedIssues);
            } catch (e) {
                setError(e.message);
            } finally {
                setLoading(false);
            }
        };
        
        if (token) {
            fetchFiringIssues();
        }
    }, [token]);

    const filteredIssues = issues.filter(
        (issue) => severityFilter === 'all' || issue.severity === severityFilter
    );

    if (loading) {
        return <div>Loading Firing Issues...</div>;
    }

    if (error) {
        return <div>Error fetching Firing Issues: {error}</div>;
    }

    return (
        <div className="firing-issues">
            <h2>Top 10 Firing Issues</h2>
            <div className="filters">
                <label htmlFor="severity-filter">Severity:</label>
                <select
                    id="severity-filter"
                    value={severityFilter}
                    onChange={(e) => setSeverityFilter(e.target.value)}
                >
                    <option value="all">All</option>
                    <option value="critical">Critical</option>
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                    <option value="page">Page</option>
                    <option value="warning">Warning</option>
                </select>
            </div>
            <ul className="issues-list">
                {filteredIssues.map((issue, index) => (
                    <li key={index} className={`issue-item ${issue.severity}`}>
                        <span className="issue-name">{issue.name}</span>
                        <span className="issue-description">{issue.description}</span>
                        <span className="issue-start-time">{issue.starts_at}</span>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default FiringIssues;
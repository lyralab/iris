import React, { useState, useEffect } from 'react';
import config from './config';

const encodedCredentials = btoa(`${config.auth.username}:${config.auth.password}`);

const ResolvedIssues = () => {
  const [resolvedSeverityFilter, setResolvedSeverityFilter] = useState('all');
  const [resolvedIssues, setResolvedIssues] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchResolvedIssues = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await fetch(config.api.resolvedIssues, {
            headers: {
              Authorization: `Basic ${encodedCredentials}`,
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
         // Calculate duration and sort by resolved time
          const issuesWithDuration = data.alerts.map((alert) => {
              const startTime = new Date(alert.starts_at);
              const endTime = new Date(alert.ends_at);
              const durationMs = endTime - startTime;
              const duration = formatDuration(durationMs);
              return {...alert, duration};
            });
          const sortedIssues = issuesWithDuration.sort((a, b) => new Date(b.ends_at) - new Date(a.ends_at)).slice(0, 10);
        setResolvedIssues(sortedIssues);
      } catch (e) {
        setError(e.message);
      } finally {
          setLoading(false);
      }
    };

    fetchResolvedIssues();
  }, []);

    const formatDuration = (durationMs) => {
        const seconds = Math.round(durationMs / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) {
            return `${days}d ${hours % 24}h`
        } else if (hours > 0) {
            return `${hours}h ${minutes % 60}m`
        }else if (minutes > 0) {
            return `${minutes}m ${seconds % 60}s`
        } else {
            return `${seconds}s`;
        }

    };

    const filteredResolvedIssues = resolvedIssues.filter(
      (issue) => resolvedSeverityFilter === 'all' || issue.severity === resolvedSeverityFilter
    );

    if (loading) {
        return <div>Loading Resolved Issues...</div>
    }

    if(error) {
        return <div>Error fetching Resolved Issues: {error}</div>
    }
  return (
    <div className="resolved-issues">
      <h2>Latest Resolved Issues</h2>
      <div className="filters">
        <label htmlFor="resolved-severity-filter">Severity:</label>
        <select
          id="resolved-severity-filter"
          value={resolvedSeverityFilter}
          onChange={(e) => setResolvedSeverityFilter(e.target.value)}
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
      <ul className="resolved-list">
        {filteredResolvedIssues.map((issue, index) => (
          <li key={index} className={`resolved-item ${issue.severity}`}>
              <div className="issue-details">
                 <span className="issue-name">{issue.name}</span>
                  <span className="issue-description">{issue.description}</span>
                  <div className="issue-times">
                      <span className="issue-start-time">Started At: {new Date(issue.starts_at).toLocaleString()}</span>
                      <span className="issue-end-time">Ended At: {new Date(issue.ends_at).toLocaleString()}</span>
                  </div>
                 <span className="issue-duration">Duration: {issue.duration}</span>
             </div>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ResolvedIssues;
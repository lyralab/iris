// import React, { useState, useEffect } from 'react';
// import config from './config';
//
//   const encodedCredentials = btoa(`${config.auth.username}:${config.auth.password}`);
//
// const AlertSummary = () => {
//     const [alerts, setAlerts] = useState({critical: 0,high: 0,medium:0,low:0, page: 0, warning: 0});
//     const [loading, setLoading] = useState(true);
//     const [error, setError] = useState(null);
//
//     useEffect(() => {
//         const fetchAlertData = async () => {
//             setLoading(true);
//              setError(null);
//
//             try {
//                 const response = await fetch(config.api.alertSummary, {
//                   headers: {
//                     Authorization: `Basic ${encodedCredentials}`,
//                     'Content-Type': 'application/json',
//                   },
//                 });
//                 if (!response.ok) {
//                     throw new Error(`HTTP error! status: ${response.status}`);
//                 }
//                 const data = await response.json();
//
//                 const mappedData = {};
//                 data.severites.forEach(item =>{
//                   mappedData[item.severity] = item.count
//                 })
//
//                 setAlerts(mappedData);
//             } catch (e) {
//                 setError(e.message)
//             } finally {
//                 setLoading(false)
//             }
//         };
//
//         fetchAlertData();
//     }, []);
//
//
//     if(loading) {
//       return <div>Loading Alert Summary...</div>
//     }
//
//     if (error) {
//       return <div>Error fetching Alert Summary: {error}</div>
//     }
//   return (
//       <div className="alert-summary">
//           <h2>Alert Summary</h2>
//           <div className="summary-items">
//               <div className="summary-item critical">
//                   <div className="icon">🔥</div>
//                   <div className="count">{alerts.critical || 0}</div>
//                   <div className="label">Critical</div>
//               </div>
//             <div className="summary-item high">
//                 <div className="icon">⚠️</div>
//                 <div className="count">{alerts.high || 0}</div>
//                 <div className="label">High</div>
//             </div>
//               <div className="summary-item medium">
//                   <div className="icon">🟡</div>
//                   <div className="count">{alerts.medium || 0}</div>
//                   <div className="label">Medium</div>
//               </div>
//               <div className="summary-item warning">
//                   <div className="icon">🟡</div>
//                   <div className="count">{alerts.warning || 0}</div>
//                   <div className="label">Warning</div>
//               </div>
//               <div className="summary-item low">
//                   <div className="icon">🟢</div>
//                   <div className="count">{alerts.low || 0}</div>
//                   <div className="label">Low</div>
//               </div>
//             <div className="summary-item page">
//                   <div className="icon">✉️</div>
//                   <div className="count">{alerts.page || 0}</div>
//                   <div className="label">Page</div>
//               </div>
//           </div>
//       </div>
//   );
// };
//
// export default AlertSummary;

import React, {useState, useEffect} from 'react';
import config from './config';

const AlertSummary = () => {
    const [alerts, setAlerts] = useState({critical: 0, high: 0, medium: 0, low: 0, page: 0, warning: 0});
    const [topAlerts, setTopAlerts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchAlertData = async () => {
            setLoading(true);
            setError(null);

            try {
                // Get the JWT token from the cookie
                const token = getCookie('jwt');

                // Fetch alert summary counts
                const summaryResponse = await fetch(config.api.alertSummary, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                });

                if (!summaryResponse.ok) {
                    throw new Error(`HTTP error! status: ${summaryResponse.status}`);
                }

                const summaryData = await summaryResponse.json();

                const mappedData = {};
                summaryData.severites.forEach(item => {
                    mappedData[item.severity] = item.count;
                });

                setAlerts(mappedData);

                // Fetch top newest critical and warning alerts
                const alertsResponse = await fetch(
                    `${config.api.alerts}?status=firing&limit=10&page=1`,
                    {
                        headers: {
                            Authorization: `Bearer ${token}`,
                            'Content-Type': 'application/json',
                        },
                    }
                );

                if (alertsResponse.ok) {
                    const alertsData = await alertsResponse.json();
                    if (alertsData.alerts && Array.isArray(alertsData.alerts)) {
                        // Filter for critical and warning severity, sort by created_at desc, take top 5
                        const filteredAlerts = alertsData.alerts
                            .filter(alert =>
                                alert.severity === 'critical' || alert.severity === 'warning'
                            )
                            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
                            .slice(0, 5);
                        setTopAlerts(filteredAlerts);
                    }
                }
            } catch (e) {
                setError(e.message);
            } finally {
                setLoading(false);
            }
        };

        fetchAlertData();
    }, []);

    // Helper function to get a cookie by name
    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return null;
    }

    if (loading) {
        return <div>Loading Alert Summary...</div>;
    }

    if (error) {
        return <div>Error fetching Alert Summary: {error}</div>;
    }

    return (
        <div className="alert-summary">
            <h2>Alert Summary</h2>
            <div className="summary-items">
                <div className="summary-item critical">
                    <div className="icon">🔥</div>
                    <div className="count">{alerts.critical || 0}</div>
                    <div className="label">Critical</div>
                </div>
                <div className="summary-item high">
                    <div className="icon">⚠️</div>
                    <div className="count">{alerts.high || 0}</div>
                    <div className="label">High</div>
                </div>
                <div className="summary-item medium">
                    <div className="icon">🟡</div>
                    <div className="count">{alerts.medium || 0}</div>
                    <div className="label">Medium</div>
                </div>
                <div className="summary-item warning">
                    <div className="icon">🟡</div>
                    <div className="count">{alerts.warning || 0}</div>
                    <div className="label">Warning</div>
                </div>
                <div className="summary-item low">
                    <div className="icon">🟢</div>
                    <div className="count">{alerts.low || 0}</div>
                    <div className="label">Low</div>
                </div>
                <div className="summary-item page">
                    <div className="icon">✉️</div>
                    <div className="count">{alerts.page || 0}</div>
                    <div className="label">Page</div>
                </div>
            </div>

            {topAlerts.length > 0 && (
                <div className="top-alerts-section">
                    <h3>Latest Critical & Warning Alerts</h3>
                    <div className="top-alerts-list">
                        {topAlerts.map((alert, index) => (
                            <div key={alert.id || index} className={`alert-item ${alert.severity}`}>
                                <div className="alert-header">
                                    <span className={`severity-badge ${alert.severity}`}>
                                        {alert.severity === 'critical' ? '🔥' : '⚠️'} {alert.severity.toUpperCase()}
                                    </span>
                                    <span className="alert-time">
                                        {new Date(alert.created_at).toLocaleString()}
                                    </span>
                                </div>
                                <div className="alert-name">{alert.name}</div>
                                <div className="alert-description">{alert.description}</div>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};


export default AlertSummary;
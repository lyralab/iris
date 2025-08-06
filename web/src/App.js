// web/src/App.js
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './LoginPage';
import AlertSummary from './AlertSummary';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/alerts" element={<AlertSummary />} />
      </Routes>
    </Router>
  );
}

export default App;
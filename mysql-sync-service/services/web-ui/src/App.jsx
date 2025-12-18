import React, { useState, useEffect } from 'react';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

function App() {
  const [status, setStatus] = useState('unknown');
  const [loading, setLoading] = useState(false);

  const fetchStatus = async () => {
    try {
      const res = await fetch(`${API_URL}/sync/status`);
      const data = await res.json();
      setStatus(data.status);
    } catch (err) {
      console.error(err);
      setStatus('error');
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
  }, []);

  const triggerSync = async () => {
    setLoading(true);
    try {
      await fetch(`${API_URL}/sync/trigger`, { method: 'POST' });
      await fetchStatus();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const stopSync = async () => {
    setLoading(true);
    try {
      await fetch(`${API_URL}/sync/stop`, { method: 'POST' });
      await fetchStatus();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>MySQL Sync Service Dashboard</h1>
      
      <div style={{ border: '1px solid #ccc', padding: '20px', borderRadius: '8px', maxWidth: '600px' }}>
        <h2>Status: <span style={{ color: status === 'running' ? 'green' : 'gray' }}>{status.toUpperCase()}</span></h2>
        
        <div style={{ marginTop: '20px' }}>
          {status === 'running' ? (
            <button 
              onClick={stopSync} 
              disabled={loading}
              style={{ padding: '10px 20px', backgroundColor: '#d9534f', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
            >
              {loading ? 'Stopping...' : 'Stop Sync'}
            </button>
          ) : (
            <button 
              onClick={triggerSync} 
              disabled={loading}
              style={{ padding: '10px 20px', backgroundColor: '#5bc0de', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
            >
              {loading ? 'Starting...' : 'Start Manual Sync'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;

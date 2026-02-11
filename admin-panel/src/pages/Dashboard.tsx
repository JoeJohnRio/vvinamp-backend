import React, { useState, useEffect } from 'react';
import { getDashboardStats } from '../utils/apiService';
import './Dashboard.css';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({
    totalTracks: 0,
    totalArtists: 0,
    totalAlbums: 0
  });

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const data = await getDashboardStats();
      setStats(data);
    } catch (error) {
      console.error('Error loading dashboard stats:', error);
    }
  };

  return (
    <div className="dashboard">
      <h1>Admin Dashboard</h1>
      <div className="stats-container">
        <div className="stat-card">
          <h3>Total Tracks</h3>
          <p>{stats.totalTracks}</p>
        </div>
        <div className="stat-card">
          <h3>Total Artists</h3>
          <p>{stats.totalArtists}</p>
        </div>
        <div className="stat-card">
          <h3>Total Albums</h3>
          <p>{stats.totalAlbums}</p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
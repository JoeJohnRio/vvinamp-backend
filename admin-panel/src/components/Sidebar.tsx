import React from 'react';
import './Sidebar.css';
import { Link } from 'react-router-dom';

const Sidebar: React.FC = () => {
  return (
    <nav className="sidebar">
      <h2>vvinamp Admin</h2>
      <ul>
        <li><Link to="/">Dashboard</Link></li>
        <li><Link to="/library">Music Library</Link></li>
        <li><Link to="/manage">Manage Music</Link></li>
        <li><Link to="/artists">Artists</Link></li>
        <li><Link to="/albums">Albums</Link></li>
        <li><Link to="/tracks">Tracks</Link></li>
      </ul>
    </nav>
  );
};

export default Sidebar;
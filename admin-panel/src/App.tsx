import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import './App.css';
import Sidebar from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import MusicLibrary from './pages/MusicLibrary';
import ManageMusic from './pages/ManageMusic';
import ArtistManagement from './pages/ArtistManagement';
import AlbumManagement from './pages/AlbumManagement';
import TrackManagement from './pages/TrackManagement';

function App() {
  return (
    <Router>
      <div className="App">
        <Sidebar />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/library" element={<MusicLibrary />} />
            <Route path="/manage" element={<ManageMusic />} />
            <Route path="/artists" element={<ArtistManagement />} />
            <Route path="/albums" element={<AlbumManagement />} />
            <Route path="/tracks" element={<TrackManagement />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;

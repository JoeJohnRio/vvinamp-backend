import React, { useState, useEffect } from 'react';
import TrackItem from '../components/TrackItem';
import FolderView from '../components/FolderView';
import UploadZone from '../components/UploadZone';
import { Track } from '../types/track';
import { getAllTracks, uploadTrack, updateTrack, deleteTrack } from '../utils/apiService';
import './MusicLibrary.css';

const MusicLibrary: React.FC = () => {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [viewMode, setViewMode] = useState<'list' | 'folders'>('list');

  useEffect(() => {
    loadTracks();
  }, []);

  const loadTracks = async () => {
    try {
      // Fetch tracks from localStorage
      const storedTracks = await getAllTracks();
      setTracks(storedTracks);
    } catch (error) {
      console.error('Error loading tracks:', error);
      // Fallback to empty array
      setTracks([]);
    }
  };

  const handleDrop = (files: File[]) => {
    // Files are handled by the UploadZone component which adds them to storage
    // We just need to refresh the track list
    loadTracks();
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTrack(id);
      loadTracks(); // Refresh after deletion
    } catch (error) {
      console.error('Error deleting track:', error);
    }
  };

  const handleDropOnTrack = async (file: File, trackId: string) => {
    // In a real implementation, this would associate the file with the specific track
    console.log(`Dropped file ${file.name} on track ${trackId}`);
    
    // For now, we'll just add the file as a new track
    const formData = new FormData();
    formData.append('file', file);
    
    try {
      await uploadTrack(formData);
      // Refresh the track list after upload
      loadTracks();
    } catch (error) {
      console.error(`Error adding ${file.name} to library:`, error);
    }
  };

  const handleDropOnAlbum = async (file: File, albumName: string) => {
    // Assign the file to the specific album
    console.log(`Dropped file ${file.name} on album ${albumName}`);
    
    // Add the file to storage and assign it to the album
    const formData = new FormData();
    formData.append('file', file);
    
    try {
      const newTrack = await uploadTrack(formData);
      // Update the track to assign it to the specific album
      await updateTrack(newTrack.id, { album: albumName });
      // Refresh the track list after upload
      loadTracks();
    } catch (error) {
      console.error(`Error adding ${file.name} to album ${albumName}:`, error);
    }
  };

  return (
    <div className="music-library">
      <div className="library-header">
        <h1>Music Library</h1>
        <div className="view-controls">
          <button 
            className={viewMode === 'list' ? 'active' : ''}
            onClick={() => setViewMode('list')}
          >
            List View
          </button>
          <button 
            className={viewMode === 'folders' ? 'active' : ''}
            onClick={() => setViewMode('folders')}
          >
            Folder View
          </button>
          <button onClick={loadTracks}>Refresh</button>
        </div>
      </div>
      
      <UploadZone onDrop={handleDrop} />
      
      <div className="current-tracks-info">
        <h2>Current Tracks in Database: {tracks.length}</h2>
      </div>
      
      {viewMode === 'list' ? (
        <div className="track-list">
          {tracks.map(track => (
            <TrackItem 
              key={track.id} 
              track={track} 
              onDelete={handleDelete}
              onDropFile={handleDropOnTrack}
            />
          ))}
        </div>
      ) : (
        <FolderView 
          tracks={tracks} 
          onDropFile={handleDropOnAlbum}
        />
      )}
    </div>
  );
};

export default MusicLibrary;
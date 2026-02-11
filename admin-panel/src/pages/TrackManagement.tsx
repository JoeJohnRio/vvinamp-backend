import React, { useState, useEffect } from 'react';
import { getAllArtists, getAllAlbums, getAllTracks, createTrack, updateTrack, deleteTrack } from '../utils/apiService';
import { Artist, Album, Track } from '../types';
import './TrackManagement.css';

const TrackManagement: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [editingTrack, setEditingTrack] = useState<Track | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    albumId: '',
    duration: 0,
    explicit: false,
    lyrics: ''
  });

  useEffect(() => {
    loadArtists();
    loadAlbums();
    loadTracks();
  }, []);

  const loadArtists = async () => {
    try {
      const data = await getAllArtists();
      setArtists(data);
    } catch (error) {
      console.error('Error loading artists:', error);
    }
  };

  const loadAlbums = async () => {
    try {
      const data = await getAllAlbums();
      setAlbums(data);
    } catch (error) {
      console.error('Error loading albums:', error);
    }
  };

  const loadTracks = async () => {
    try {
      const data = await getAllTracks();
      setTracks(data);
    } catch (error) {
      console.error('Error loading tracks:', error);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    const checked = type === 'checkbox' ? (e.target as HTMLInputElement).checked : undefined;
    
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : (type === 'number' ? Number(value) : value)
    }));
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const formDataToSend = new FormData();
      
      if (selectedFile) {
        formDataToSend.append('file', selectedFile);
      }
      
      Object.keys(formData).forEach(key => {
        if (key !== 'file') {
          formDataToSend.append(key, String((formData as any)[key]));
        }
      });

      if (editingTrack) {
        // For updates, we might need a different approach since we can't send file with other data in multipart
        // For now, we'll just update metadata
        await updateTrack(editingTrack.id, formData);
      } else {
        // Create new track with file
        await createTrack(formDataToSend);
      }
      
      resetForm();
      loadTracks();
    } catch (error) {
      console.error('Error saving track:', error);
    }
  };

  const handleEdit = (track: Track) => {
    setEditingTrack(track);
    setFormData({
      title: track.title || '',
      albumId: '', // We'll need to find the album ID based on the track's album
      duration: track.duration || 0,
      explicit: track.explicit || false,
      lyrics: track.lyrics || ''
    });
    
    // Find the album ID based on the track's album
    const album = albums.find(a => a.title === track.album);
    if (album) {
      setFormData(prev => ({ ...prev, albumId: album.id }));
    }
    
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this track?')) {
      try {
        await deleteTrack(id);
        loadTracks();
      } catch (error) {
        console.error('Error deleting track:', error);
      }
    }
  };

  const resetForm = () => {
    setEditingTrack(null);
    setFormData({
      title: '',
      albumId: '',
      duration: 0,
      explicit: false,
      lyrics: ''
    });
    setSelectedFile(null);
    setShowForm(false);
  };

  return (
    <div className="track-management">
      <h1>Track Management</h1>
      
      <div className="controls">
        <button onClick={() => {resetForm(); setShowForm(true);}} className="btn-primary">
          Add New Track
        </button>
      </div>

      {showForm && (
        <div className="form-container">
          <h2>{editingTrack ? 'Edit Track' : 'Add New Track'}</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="title">Track Title *</label>
              <input
                type="text"
                id="title"
                name="title"
                value={formData.title}
                onChange={handleInputChange}
                required
              />
            </div>
            
            {!editingTrack && (
              <>
                <div className="form-group">
                  <label htmlFor="file">Audio File *</label>
                  <input
                    type="file"
                    id="file"
                    name="file"
                    accept="audio/mp3,audio/wav,audio/flac,audio/mpeg"
                    onChange={handleFileChange}
                    required={!editingTrack}
                  />
                </div>
                
                <div className="form-group">
                  <label htmlFor="albumId">Album *</label>
                  <select
                    id="albumId"
                    name="albumId"
                    value={formData.albumId}
                    onChange={handleInputChange}
                    required
                  >
                    <option value="">Select an album</option>
                    {albums.map(album => (
                      <option key={album.id} value={album.id}>
                        {album.title} ({album.artistName})
                      </option>
                    ))}
                  </select>
                </div>
              </>
            )}
            
            <div className="form-group">
              <label htmlFor="duration">Duration (seconds)</label>
              <input
                type="number"
                id="duration"
                name="duration"
                value={formData.duration}
                onChange={handleInputChange}
                min="0"
              />
            </div>
            
            <div className="form-group checkbox-group">
              <label htmlFor="explicit">Explicit Content</label>
              <input
                type="checkbox"
                id="explicit"
                name="explicit"
                checked={formData.explicit}
                onChange={handleInputChange}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="lyrics">Lyrics</label>
              <textarea
                id="lyrics"
                name="lyrics"
                value={formData.lyrics}
                onChange={handleInputChange}
                rows={6}
              />
            </div>
            
            <div className="form-actions">
              <button type="submit" className="btn-primary">
                {editingTrack ? 'Update Track' : 'Create Track'}
              </button>
              <button type="button" onClick={resetForm} className="btn-secondary">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="items-grid">
        {tracks.map(track => (
          <div key={track.id} className="item-card">
            <div className="card-header">
              <h3>{track.title}</h3>
              <div className="card-actions">
                <button onClick={() => handleEdit(track)} className="btn-edit">Edit</button>
                <button onClick={() => handleDelete(track.id)} className="btn-delete">Delete</button>
              </div>
            </div>
            <div className="card-body">
              <p><strong>Artist:</strong> {track.artist}</p>
              <p><strong>Album:</strong> {track.album}</p>
              <p><strong>Duration:</strong> {Math.floor(track.duration / 60)}:{String(track.duration % 60).padStart(2, '0')}</p>
              <p><strong>Explicit:</strong> {track.explicit ? 'Yes' : 'No'}</p>
              {track.fileName && <p><strong>File:</strong> {track.fileName.split('/').pop()}</p>}
            </div>
            <div className="card-footer">
              <small>Uploaded: {track.uploadDate ? new Date(track.uploadDate).toLocaleDateString() : 'N/A'}</small>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TrackManagement;
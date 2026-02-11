import React, { useState, useEffect } from 'react';
import { getAllArtists, getAllAlbums, createAlbum, updateAlbum, deleteAlbum } from '../utils/apiService';
import { Artist, Album } from '../types';
import './AlbumManagement.css';

const AlbumManagement: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [editingAlbum, setEditingAlbum] = useState<Album | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    artistId: '',
    releaseDate: '',
    coverArt: ''
  });

  useEffect(() => {
    loadArtists();
    loadAlbums();
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

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      if (editingAlbum) {
        // Update existing album
        await updateAlbum(editingAlbum.id, formData);
      } else {
        // Create new album
        await createAlbum(formData);
      }
      
      resetForm();
      loadAlbums();
    } catch (error) {
      console.error('Error saving album:', error);
    }
  };

  const handleEdit = (album: Album) => {
    setEditingAlbum(album);
    setFormData({
      title: album.title || '',
      artistId: album.artistId || '',
      releaseDate: album.releaseDate || '',
      coverArt: album.coverArt || ''
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this album? This will also remove all associated tracks.')) {
      try {
        await deleteAlbum(id);
        loadAlbums();
      } catch (error) {
        console.error('Error deleting album:', error);
      }
    }
  };

  const resetForm = () => {
    setEditingAlbum(null);
    setFormData({
      title: '',
      artistId: '',
      releaseDate: '',
      coverArt: ''
    });
    setShowForm(false);
  };

  return (
    <div className="album-management">
      <h1>Album Management</h1>
      
      <div className="controls">
        <button onClick={() => {resetForm(); setShowForm(true);}} className="btn-primary">
          Add New Album
        </button>
      </div>

      {showForm && (
        <div className="form-container">
          <h2>{editingAlbum ? 'Edit Album' : 'Add New Album'}</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="title">Album Title *</label>
              <input
                type="text"
                id="title"
                name="title"
                value={formData.title}
                onChange={handleInputChange}
                required
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="artistId">Artist *</label>
              <select
                id="artistId"
                name="artistId"
                value={formData.artistId}
                onChange={handleInputChange}
                required
              >
                <option value="">Select an artist</option>
                {artists.map(artist => (
                  <option key={artist.id} value={artist.id}>
                    {artist.name}
                  </option>
                ))}
              </select>
            </div>
            
            <div className="form-group">
              <label htmlFor="releaseDate">Release Date</label>
              <input
                type="date"
                id="releaseDate"
                name="releaseDate"
                value={formData.releaseDate}
                onChange={handleInputChange}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="coverArt">Cover Art URL</label>
              <input
                type="text"
                id="coverArt"
                name="coverArt"
                value={formData.coverArt}
                onChange={handleInputChange}
              />
            </div>
            
            <div className="form-actions">
              <button type="submit" className="btn-primary">
                {editingAlbum ? 'Update Album' : 'Create Album'}
              </button>
              <button type="button" onClick={resetForm} className="btn-secondary">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="items-grid">
        {albums.map(album => (
          <div key={album.id} className="item-card">
            <div className="card-header">
              <h3>{album.title}</h3>
              <div className="card-actions">
                <button onClick={() => handleEdit(album)} className="btn-edit">Edit</button>
                <button onClick={() => handleDelete(album.id)} className="btn-delete">Delete</button>
              </div>
            </div>
            <div className="card-body">
              <p><strong>Artist:</strong> {album.artistName}</p>
              <p><strong>Release Date:</strong> {album.releaseDate ? new Date(album.releaseDate).toLocaleDateString() : 'Not set'}</p>
              {album.coverArt && <p><strong>Cover Art:</strong> Available</p>}
            </div>
            <div className="card-footer">
              <small>Created: {album.createdAt ? new Date(album.createdAt).toLocaleDateString() : 'N/A'}</small>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default AlbumManagement;
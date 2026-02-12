import React, { useState, useEffect, useRef } from 'react';
import { getAllArtists, getAllAlbums, createAlbum, updateAlbum, deleteAlbum } from '../utils/apiService';
import { Artist, Album } from '../types';
import './AlbumManagement.css';

const AlbumManagement: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [editingAlbum, setEditingAlbum] = useState<Album | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    artistId: '',
    releaseDate: '',
    coverArt: ''
  });
  const fileInputRef = useRef<HTMLInputElement>(null);

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

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setSelectedImage(file);

      // Create a preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];

      // Validate file type
      if (!file.type.match('image.*')) {
        alert('Please drop an image file');
        return;
      }

      setSelectedImage(file);

      // Create a preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const formDataToSend = new FormData();

      // Add form fields
      Object.keys(formData).forEach(key => {
        formDataToSend.append(key, String((formData as any)[key]));
      });

      // Add image if selected
      if (selectedImage) {
        formDataToSend.append('coverArt', selectedImage);
      }

      if (editingAlbum) {
        // For updates, send the complete FormData (with image if selected)
        await updateAlbum(editingAlbum.id, formDataToSend);
      } else {
        // Create new album with image
        await createAlbum(formDataToSend);
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
    // Set the image preview if the album has cover art
    setImagePreview(album.coverArt || null);
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
    setSelectedImage(null);
    setImagePreview(null);
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
            
            <div
              className={`form-group drop-area ${imagePreview ? 'has-image' : ''}`}
              onDrop={handleDrop}
              onDragOver={handleDragOver}
              onClick={() => fileInputRef.current?.click()}
            >
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleImageChange}
                style={{ display: 'none' }}
                accept="image/*"
              />
              {imagePreview ? (
                <img src={imagePreview} alt="Preview" className="image-preview" />
              ) : (
                <div className="drop-message">
                  <p>Drag & drop album cover image here or click to browse</p>
                  <span>Supports JPG, PNG, GIF</span>
                </div>
              )}
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
              {album.coverArt && (
                <img src={album.coverArt} alt={`${album.title} cover`} className="album-cover" style={{ maxWidth: '100%', height: 'auto', marginBottom: '10px' }} />
              )}
              <p><strong>Artist:</strong> {album.artistName}</p>
              <p><strong>Release Date:</strong> {album.releaseDate ? new Date(album.releaseDate).toLocaleDateString() : 'Not set'}</p>
            </div>
            <div className="card-footer">
              <small>Created: {album.releaseDate ? new Date(album.releaseDate).toLocaleDateString() : 'N/A'}</small>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default AlbumManagement;
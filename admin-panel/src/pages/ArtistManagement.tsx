import React, { useState, useEffect, useRef } from 'react';
import { getAllArtists, createArtist, updateArtist, deleteArtist } from '../utils/apiService';
import { Artist } from '../types';
import './ArtistManagement.css';

const ArtistManagement: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [editingArtist, setEditingArtist] = useState<Artist | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    bio: '',
    verified: false
  });
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadArtists();
  }, []);

  const loadArtists = async () => {
    try {
      const data = await getAllArtists();
      setArtists(data);
    } catch (error) {
      console.error('Error loading artists:', error);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    const checked = type === 'checkbox' ? (e.target as HTMLInputElement).checked : undefined;
    
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
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
        formDataToSend.append('profileImage', selectedImage);
      }
      
      if (editingArtist) {
        // For updates, send the complete FormData (with image if selected)
        await updateArtist(editingArtist.id, formDataToSend);
      } else {
        // Create new artist with image
        await createArtist(formDataToSend);
      }
      
      resetForm();
      loadArtists();
    } catch (error) {
      console.error('Error saving artist:', error);
    }
  };

  const handleEdit = (artist: Artist) => {
    setEditingArtist(artist);
    setFormData({
      name: artist.name || '',
      bio: artist.bio || '',
      verified: artist.verified || false
    });
    setImagePreview(artist.profileImage || null);
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (window.confirm('Are you sure you want to delete this artist?')) {
      try {
        await deleteArtist(id);
        loadArtists();
      } catch (error) {
        console.error('Error deleting artist:', error);
      }
    }
  };

  const resetForm = () => {
    setEditingArtist(null);
    setFormData({
      name: '',
      bio: '',
      verified: false
    });
    setSelectedImage(null);
    setImagePreview(null);
    setShowForm(false);
  };

  return (
    <div className="artist-management">
      <h1>Artist Management</h1>
      
      <div className="controls">
        <button onClick={() => {resetForm(); setShowForm(true);}} className="btn-primary">
          Add New Artist
        </button>
      </div>

      {showForm && (
        <div className="form-container">
          <h2>{editingArtist ? 'Edit Artist' : 'Add New Artist'}</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="name">Artist Name *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="bio">Biography</label>
              <textarea
                id="bio"
                name="bio"
                value={formData.bio}
                onChange={handleInputChange}
                rows={4}
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
                  <p>Drag & drop artist image here or click to browse</p>
                  <span>Supports JPG, PNG, GIF</span>
                </div>
              )}
            </div>
            
            <div className="form-group checkbox-group">
              <label htmlFor="verified">Verified Artist</label>
              <input
                type="checkbox"
                id="verified"
                name="verified"
                checked={formData.verified}
                onChange={handleInputChange}
              />
            </div>
            
            <div className="form-actions">
              <button type="submit" className="btn-primary">
                {editingArtist ? 'Update Artist' : 'Create Artist'}
              </button>
              <button type="button" onClick={resetForm} className="btn-secondary">
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="items-grid">
        {artists.map(artist => (
          <div key={artist.id} className="item-card">
            <div className="card-header">
              <h3>{artist.name}</h3>
              <div className="card-actions">
                <button onClick={() => handleEdit(artist)} className="btn-edit">Edit</button>
                <button onClick={() => handleDelete(artist.id)} className="btn-delete">Delete</button>
              </div>
            </div>
            <div className="card-body">
              {artist.profileImage && (
                <img src={artist.profileImage} alt={artist.name} className="artist-image" />
              )}
              <p><strong>Verified:</strong> {artist.verified ? 'Yes' : 'No'}</p>
              <p><strong>Monthly Listeners:</strong> {artist.monthlyListeners || 0}</p>
              {artist.bio && <p><strong>Bio:</strong> {artist.bio}</p>}
            </div>
            <div className="card-footer">
              <small>Created: {artist.createdAt ? new Date(artist.createdAt).toLocaleDateString() : 'N/A'}</small>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ArtistManagement;
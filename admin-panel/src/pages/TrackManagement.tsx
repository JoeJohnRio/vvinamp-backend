import React, { useState, useEffect, useRef } from 'react';
import { getAllArtists, getAllAlbums, getAllTracks, createTrack, updateTrack, updateTrackFile, deleteTrack } from '../utils/apiService';
import { Artist, Album, Track } from '../types';
import './TrackManagement.css';

const TrackManagement: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [editingTrack, setEditingTrack] = useState<Track | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [filePreview, setFilePreview] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    albumId: '',
    duration: 0,
    explicit: false,
    lyrics: ''
  });
  const fileInputRef = useRef<HTMLInputElement>(null);

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
      const file = e.target.files[0];
      setSelectedFile(file);

      // Create a preview URL for the audio file
      const reader = new FileReader();
      reader.onloadend = () => {
        setFilePreview(reader.result as string);
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
      if (!file.type.startsWith('audio/')) {
        alert('Please drop an audio file (mp3, wav, flac, mpeg)');
        return;
      }

      setSelectedFile(file);

      // Create a preview URL for the audio file
      const reader = new FileReader();
      reader.onloadend = () => {
        setFilePreview(reader.result as string);
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
      if (editingTrack) {
        // For updates, handle differently based on whether a new file was selected
        if (selectedFile) {
          // If a new file was selected, update the track file
          const formDataToSend = new FormData();
          formDataToSend.append('file', selectedFile);
          
          // Add other form data
          Object.keys(formData).forEach(key => {
            if (key !== 'file') {
              formDataToSend.append(key, String((formData as any)[key]));
            }
          });
          
          await updateTrackFile(editingTrack.id, formDataToSend);
        } else {
          // Just update metadata
          await updateTrack(editingTrack.id, formData);
        }
      } else {
        // Create new track with file
        const formDataToSend = new FormData();
        
        if (selectedFile) {
          formDataToSend.append('file', selectedFile);
        }
        
        Object.keys(formData).forEach(key => {
          if (key !== 'file') {
            formDataToSend.append(key, String((formData as any)[key]));
          }
        });
        
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
    
    // Find the album ID based on the track's album
    const album = albums.find(a => a.title === track.album);
    const albumId = album ? album.id : '';

    setFormData({
      title: track.title || '',
      albumId: albumId,
      duration: track.duration || 0,
      explicit: track.explicit || false,
      lyrics: track.lyrics || ''
    });

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
    setFilePreview(null);
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

            {(editingTrack ? true : true) && (
              <>
                <div
                  className={`form-group drop-area ${filePreview ? 'has-file' : ''}`}
                  onDrop={handleDrop}
                  onDragOver={handleDragOver}
                  onClick={() => fileInputRef.current?.click()}
                >
                  <input
                    type="file"
                    ref={fileInputRef}
                    onChange={handleFileChange}
                    style={{ display: 'none' }}
                    accept="audio/*"
                  />
                  {filePreview ? (
                    <div className="file-preview">
                      <p>Selected: {selectedFile?.name}</p>
                      <audio controls src={filePreview} className="audio-preview" />
                    </div>
                  ) : (
                    editingTrack && !selectedFile ? (
                      <div className="drop-message">
                        <p>Drag & drop new audio file here to replace current file, or click to browse</p>
                        <span>Supports MP3, WAV, FLAC, MPEG</span>
                      </div>
                    ) : (
                      <div className="drop-message">
                        <p>Drag & drop audio file here or click to browse</p>
                        <span>Supports MP3, WAV, FLAC, MPEG</span>
                      </div>
                    )
                  )}
                </div>

                <div className="form-group">
                  <label htmlFor="albumId">Album *</label>
                  <select
                    id="albumId"
                    name="albumId"
                    value={formData.albumId}
                    onChange={handleInputChange}
                    required={!editingTrack} // Not required when editing since it's already set
                  >
                    <option value="">Select an album</option>
                    {albums.map(album => (
                      <option key={album.id} value={album.id}>
                        {album.title} ({album.artistName})
                      </option>
                    ))}
                  </select>
                </div>
                
                {/* Show current file info when editing and no new file selected */}
                {editingTrack && !selectedFile && (
                  <div className="form-group">
                    <label>Current File</label>
                    <div className="current-file-info">
                      <p>{editingTrack.fileName ? editingTrack.fileName.split('/').pop() : 'No file available'}</p>
                    </div>
                  </div>
                )}
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
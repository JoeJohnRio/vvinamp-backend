import React, { useState, useEffect } from 'react';
import { getAllTracks, getAllArtists, getAllAlbums, createArtist, createAlbum, createTrack } from '../utils/apiService';
import { Track, Artist, Album } from '../types';
import './ManageMusic.css';

const ManageMusic: React.FC = () => {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [selectedArtist, setSelectedArtist] = useState<string>('');
  const [selectedAlbum, setSelectedAlbum] = useState<string>('');
  const [newArtistName, setNewArtistName] = useState('');
  const [newAlbumTitle, setNewAlbumTitle] = useState('');
  const [newTrackTitle, setNewTrackTitle] = useState('');
  const [newTrackFile, setNewTrackFile] = useState<File | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [loadedArtists, loadedAlbums, loadedTracks] = await Promise.all([
        getAllArtists(),
        getAllAlbums(),
        getAllTracks()
      ]);
      
      setArtists(loadedArtists);
      setAlbums(loadedAlbums);
      setTracks(loadedTracks);
    } catch (error) {
      console.error('Error loading data:', error);
    }
  };

  const handleCreateArtist = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newArtistName.trim()) return;

    try {
      await createArtist({ name: newArtistName });
      setNewArtistName('');
      loadData(); // Refresh the data
    } catch (error) {
      console.error('Error creating artist:', error);
    }
  };

  const handleCreateAlbum = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newAlbumTitle.trim() || !selectedArtist) return;

    try {
      await createAlbum({ title: newAlbumTitle, artistId: selectedArtist });
      setNewAlbumTitle('');
      loadData(); // Refresh the data
    } catch (error) {
      console.error('Error creating album:', error);
    }
  };

  const handleCreateTrack = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTrackTitle.trim() || !selectedAlbum || !newTrackFile) return;

    try {
      const formData = new FormData();
      formData.append('file', newTrackFile);
      formData.append('title', newTrackTitle);
      formData.append('albumId', selectedAlbum);

      await createTrack(formData);
      setNewTrackTitle('');
      setNewTrackFile(null);
      loadData(); // Refresh the data
    } catch (error) {
      console.error('Error creating track:', error);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setNewTrackFile(e.target.files[0]);
    }
  };

  return (
    <div className="manage-music">
      <h1>Manage Music Library</h1>

      {/* Artist Management */}
      <section className="management-section">
        <h2>Artists</h2>
        <form onSubmit={handleCreateArtist} className="create-form">
          <input
            type="text"
            value={newArtistName}
            onChange={(e) => setNewArtistName(e.target.value)}
            placeholder="New Artist Name"
            required
          />
          <button type="submit">Add Artist</button>
        </form>
        
        <div className="items-list">
          {artists.map(artist => (
            <div key={artist.id} className="item-card">
              <h3>{artist.name}</h3>
              <p>ID: {artist.id}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Album Management */}
      <section className="management-section">
        <h2>Albums</h2>
        <form onSubmit={handleCreateAlbum} className="create-form">
          <select 
            value={selectedArtist} 
            onChange={(e) => setSelectedArtist(e.target.value)}
            required
          >
            <option value="">Select Artist</option>
            {artists.map(artist => (
              <option key={artist.id} value={artist.id}>{artist.name}</option>
            ))}
          </select>
          <input
            type="text"
            value={newAlbumTitle}
            onChange={(e) => setNewAlbumTitle(e.target.value)}
            placeholder="New Album Title"
            required
          />
          <button type="submit">Add Album</button>
        </form>
        
        <div className="items-list">
          {albums.map(album => (
            <div key={album.id} className="item-card">
              <h3>{album.title}</h3>
              <p>Artist: {album.artistName}</p>
              <p>ID: {album.id}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Track Management */}
      <section className="management-section">
        <h2>Tracks</h2>
        <form onSubmit={handleCreateTrack} className="create-form">
          <select 
            value={selectedAlbum} 
            onChange={(e) => setSelectedAlbum(e.target.value)}
            required
          >
            <option value="">Select Album</option>
            {albums.map(album => (
              <option key={album.id} value={album.id}>{album.title} ({album.artistName})</option>
            ))}
          </select>
          <input
            type="text"
            value={newTrackTitle}
            onChange={(e) => setNewTrackTitle(e.target.value)}
            placeholder="New Track Title"
            required
          />
          <input
            type="file"
            accept="audio/mp3,audio/wav,audio/flac,audio/mpeg"
            onChange={handleFileChange}
            required
          />
          <button type="submit">Add Track</button>
        </form>
        
        <div className="items-list">
          {tracks.map(track => (
            <div key={track.id} className="item-card">
              <h3>{track.title}</h3>
              <p>Artist: {track.artist}</p>
              <p>Album: {track.album}</p>
              <p>Duration: {track.duration}s</p>
              <p>File: {track.fileName}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
};

export default ManageMusic;
import React from 'react';
import AlbumFolder from './AlbumFolder';
import ArtistFolder from './ArtistFolder';
import { Track } from '../types/track';
import './FolderView.css';

interface FolderViewProps {
  tracks: Track[];
  onDropFile?: (file: File, target: string) => void;
}

const FolderView: React.FC<FolderViewProps> = ({ tracks, onDropFile }) => {
  // Group tracks by artist
  const groupedTracks: Record<string, Track[]> = {};
  tracks.forEach(track => {
    if (!groupedTracks[track.artist]) {
      groupedTracks[track.artist] = [];
    }
    groupedTracks[track.artist].push(track);
  });

  // Group tracks by album as well
  const groupedByAlbum: Record<string, Track[]> = {};
  tracks.forEach(track => {
    if (!groupedByAlbum[track.album]) {
      groupedByAlbum[track.album] = [];
    }
    groupedByAlbum[track.album].push(track);
  });

  // Handle dropping files anywhere in the folder view
  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.currentTarget.classList.add('drop-target');
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.currentTarget.classList.remove('drop-target');
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.currentTarget.classList.remove('drop-target');
    
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0 && onDropFile) {
      // For general drops, we could implement a default behavior
      // For now, we'll just log that a general drop occurred
      console.log('General drop on folder view');
    }
  };

  return (
    <div 
      className="folder-view"
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div className="albums-section">
        <h3>Albums</h3>
        {Object.entries(groupedByAlbum).map(([album, albumTracks]) => (
          <AlbumFolder 
            key={album} 
            album={album} 
            tracks={albumTracks} 
            onDropFile={onDropFile} 
          />
        ))}
      </div>

      <div className="artists-section">
        <h3>Artists</h3>
        {Object.entries(groupedTracks).map(([artist, artistTracks]) => (
          <ArtistFolder 
            key={artist} 
            artist={artist} 
            tracks={artistTracks} 
            onDropFile={onDropFile} 
          />
        ))}
      </div>
    </div>
  );
};

export default FolderView;
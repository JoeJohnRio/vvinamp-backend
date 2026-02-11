import React, { useState } from 'react';
import { Track } from '../types/track';
import './ArtistFolder.css';

interface ArtistFolderProps {
  artist: string;
  tracks: Track[];
  onDropFile?: (file: File, targetArtist: string) => void;
}

const ArtistFolder: React.FC<ArtistFolderProps> = ({ artist, tracks, onDropFile }) => {
  const [expanded, setExpanded] = useState(false);
  
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
      // Get the files that were dragged
      Array.from(e.dataTransfer.files).forEach(file => {
        onDropFile(file, artist);
      });
    }
  };

  const toggleFolder = () => {
    setExpanded(!expanded);
  };

  return (
    <div 
      className="folder"
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div 
        className="folder-header" 
        onClick={toggleFolder}
      >
        <span className="folder-icon">
          {expanded ? '📂' : '📁'}
        </span>
        <span className="folder-name">{artist}</span>
        <span className="folder-count">({tracks.length})</span>
      </div>
      
      {expanded && (
        <div className="folder-content">
          {tracks.map(track => (
            <div key={track.id} className="folder-track">
              <span className="track-title">{track.title}</span>
              <span className="track-album">{track.album}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ArtistFolder;
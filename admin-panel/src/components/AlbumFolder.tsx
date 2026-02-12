import React, { useState } from 'react';
import { Track } from '../types/track';
import './AlbumFolder.css';

interface AlbumFolderProps {
  album: string;
  tracks: Track[];
  onDropFile?: (file: File, targetAlbum: string) => void;
}

const AlbumFolder: React.FC<AlbumFolderProps> = ({ album, tracks, onDropFile }) => {
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
        onDropFile(file, album);
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
          {expanded ? '📀' : '💿'}
        </span>
        <span className="folder-name">{album}</span>
        <span className="folder-count">({tracks.length})</span>
      </div>

      {expanded && (
        <div className="folder-content">
          {tracks.map(track => (
            <div key={track.id} className="folder-track">
              <span className="track-title">{track.title}</span>
              <span className="track-artist">{track.artist}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AlbumFolder;
import React from 'react';
import { Track } from '../types/track';
import './AlbumFolder.css';

interface AlbumFolderProps {
  album: string;
  tracks: Track[];
  onDropFile?: (file: File, targetAlbum: string) => void;
}

const AlbumFolder: React.FC<AlbumFolderProps> = ({ album, tracks, onDropFile }) => {
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

  return (
    <div 
      className="album-folder"
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div className="album-header">
        <span className="folder-icon">💿</span>
        <span className="album-name">{album}</span>
        <span className="album-count">({tracks.length})</span>
      </div>
    </div>
  );
};

export default AlbumFolder;
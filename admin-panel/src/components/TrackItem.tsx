import React from 'react';
import { Track } from '../types/track';
import { deleteTrack } from '../utils/apiService';
import './TrackItem.css';

interface TrackItemProps {
  track: Track;
  onDelete?: (id: string) => void;
  onDropFile?: (file: File, trackId: string) => void;
}

const TrackItem: React.FC<TrackItemProps> = ({ track, onDelete, onDropFile }) => {
  // For now, we'll implement a simple drag and drop handler without react-dnd
  // since we're having issues with the DndProvider
  
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
      // Get the file that was dragged
      const file = e.dataTransfer.files[0];
      onDropFile(file, track.id);
    }
  };

  const handleDelete = async () => {
    try {
      await deleteTrack(track.id);
      if (onDelete) {
        onDelete(track.id);
      }
    } catch (error) {
      console.error('Error deleting track:', error);
    }
  };

  return (
    <div 
      className="track-item"
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div className="track-info">
        <h4>{track.title}</h4>
        <p>{track.artist} • {track.album}</p>
      </div>
      <div className="track-actions">
        <button className="btn-edit">Edit</button>
        <button className="btn-delete" onClick={handleDelete}>Delete</button>
      </div>
    </div>
  );
};

export default TrackItem;
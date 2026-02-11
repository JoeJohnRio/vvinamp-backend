import React, { useRef, useState } from 'react';
import { uploadTrack } from '../utils/apiService';
import './UploadZone.css';

interface UploadZoneProps {
  onDrop: (files: File[]) => void;
}

const UploadZone: React.FC<UploadZoneProps> = ({ onDrop }) => {
  const [isOver, setIsOver] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFiles = (files: File[]) => {
    // Process files locally first
    onDrop(files);
    
    // Then upload to storage
    uploadFilesToStorage(files);
  };

  const uploadFilesToStorage = async (files: File[]) => {
    setIsUploading(true);
    
    for (const file of files) {
      if (file.type.startsWith('audio/')) {
        const formData = new FormData();
        formData.append('file', file);
        
        try {
          await uploadTrack(formData);
          console.log(`Successfully added ${file.name} to storage`);
        } catch (error) {
          console.error(`Error adding ${file.name} to storage:`, error);
        }
      }
    }
    
    setIsUploading(false);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      handleFiles(Array.from(e.target.files));
    }
  };

  const handleClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsOver(true);
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsOver(false);
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsOver(false);
    
    if (e.dataTransfer.files) {
      handleFiles(Array.from(e.dataTransfer.files));
    }
  };

  return (
    <div 
      className={`upload-zone ${isOver ? 'drag-over' : ''} ${isUploading ? 'uploading' : ''}`}
      onClick={handleClick}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileSelect}
        style={{ display: 'none' }}
        multiple
        accept="audio/*"
      />
      <div className="upload-content">
        <p>Drag & drop audio files here or click to browse</p>
        <span>Supported formats: MP3, WAV, FLAC, etc.</span>
        {isUploading && <p className="upload-status">Adding to library...</p>}
      </div>
    </div>
  );
};

export default UploadZone;
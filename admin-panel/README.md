# vvinamp Admin Panel

This is an admin web interface for managing your vvinamp music library. It allows you to upload audio files, organize them into folders, and manage metadata.

## Features

- File uploads (via drag and drop or file browser)
- Folder organization view
- Track management (edit, delete)
- Responsive design

## Prerequisites

- Node.js (v14 or higher)
- Your Go backend server running

## Setup

1. Install dependencies:
```bash
npm install
```

2. Create a `.env` file in the root of this project with your backend API URL:
```env
REACT_APP_API_URL=http://localhost:8081
```

3. Start the development server:
```bash
npm start
```

The admin panel will be accessible at `http://localhost:3001`.

## API Integration

This admin panel is designed to connect to your existing Go backend. You'll need to implement the following API endpoints in your Go application:

- `GET /api/tracks` - Retrieve all tracks
- `POST /api/tracks/upload` - Upload a new track
- `DELETE /api/tracks/:id` - Delete a track
- `PUT /api/tracks/:id` - Update track metadata

## Usage

1. Navigate to the Music Library section
2. Drag and drop audio files into the upload zone or click to browse
3. Organize tracks using the list or folder views
4. Edit or delete tracks as needed

## Tech Stack

- React with TypeScript
- React DnD for drag and drop functionality
- Axios for API calls
- CSS for styling
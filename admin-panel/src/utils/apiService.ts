// apiService.ts
import axios from 'axios';

// Get the backend API URL from environment variable
const BACKEND_API_URL = process.env.REACT_APP_BACKEND_API_URL || 'http://localhost:8081/api';

// Track-related functions
// Function to upload a track
export const uploadTrack = async (formData: FormData) => {
  try {
    const response = await axios.post(`${BACKEND_API_URL}/tracks/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return response.data;
  } catch (error) {
    console.error('Error uploading track:', error);
    throw error;
  }
};

// Function to get all tracks
export const getAllTracks = async () => {
  try {
    const response = await axios.get(`${BACKEND_API_URL}/tracks`);
    return response.data;
  } catch (error) {
    console.error('Error fetching tracks:', error);
    return [];
  }
};

// Function to update a track
export const updateTrack = async (trackId: string, trackData: any) => {
  try {
    const response = await axios.put(`${BACKEND_API_URL}/tracks/${trackId}`, trackData);
    return response.data;
  } catch (error) {
    console.error('Error updating track:', error);
    throw error;
  }
};

// Function to delete a track
export const deleteTrack = async (trackId: string) => {
  try {
    const response = await axios.delete(`${BACKEND_API_URL}/tracks/${trackId}`);
    return response.data;
  } catch (error) {
    console.error('Error deleting track:', error);
    throw error;
  }
};

// Dashboard functions
// Function to get dashboard statistics
export const getDashboardStats = async () => {
  try {
    const response = await axios.get(`${BACKEND_API_URL}/stats`);
    return response.data;
  } catch (error) {
    console.error('Error fetching dashboard stats:', error);
    return { totalTracks: 0, totalArtists: 0, totalAlbums: 0 };
  }
};

// Artist-related functions
// Function to get all artists
export const getAllArtists = async () => {
  try {
    const response = await axios.get(`${BACKEND_API_URL}/artists`);
    return response.data;
  } catch (error) {
    console.error('Error fetching artists:', error);
    return [];
  }
};

// Function to create a new artist
export const createArtist = async (artistData: any) => {
  try {
    // Check if artistData is FormData (for image uploads) or regular object
    let response;
    if (artistData instanceof FormData) {
      response = await axios.post(`${BACKEND_API_URL}/artists`, artistData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    } else {
      response = await axios.post(`${BACKEND_API_URL}/artists`, artistData);
    }
    return response.data;
  } catch (error) {
    console.error('Error creating artist:', error);
    throw error;
  }
};

// Function to update an artist
export const updateArtist = async (artistId: string, artistData: any) => {
  try {
    // Check if artistData is FormData (for image uploads) or regular object
    let response;
    if (artistData instanceof FormData) {
      response = await axios.put(`${BACKEND_API_URL}/artists/${artistId}`, artistData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    } else {
      response = await axios.put(`${BACKEND_API_URL}/artists/${artistId}`, artistData);
    }
    return response.data;
  } catch (error) {
    console.error('Error updating artist:', error);
    throw error;
  }
};

// Function to delete an artist
export const deleteArtist = async (artistId: string) => {
  try {
    const response = await axios.delete(`${BACKEND_API_URL}/artists/${artistId}`);
    return response.data;
  } catch (error) {
    console.error('Error deleting artist:', error);
    throw error;
  }
};

// Album-related functions
// Function to get all albums
export const getAllAlbums = async () => {
  try {
    const response = await axios.get(`${BACKEND_API_URL}/albums`);
    return response.data;
  } catch (error) {
    console.error('Error fetching albums:', error);
    return [];
  }
};

// Function to create a new album
export const createAlbum = async (albumData: any) => {
  try {
    // Check if albumData is FormData (for image uploads) or regular object
    let response;
    if (albumData instanceof FormData) {
      response = await axios.post(`${BACKEND_API_URL}/albums`, albumData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    } else {
      response = await axios.post(`${BACKEND_API_URL}/albums`, albumData);
    }
    return response.data;
  } catch (error) {
    console.error('Error creating album:', error);
    throw error;
  }
};

// Function to update an album
export const updateAlbum = async (albumId: string, albumData: any) => {
  try {
    // Check if albumData is FormData (for image uploads) or regular object
    let response;
    if (albumData instanceof FormData) {
      response = await axios.put(`${BACKEND_API_URL}/albums/${albumId}`, albumData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    } else {
      response = await axios.put(`${BACKEND_API_URL}/albums/${albumId}`, albumData);
    }
    return response.data;
  } catch (error) {
    console.error('Error updating album:', error);
    throw error;
  }
};

// Function to delete an album
export const deleteAlbum = async (albumId: string) => {
  try {
    const response = await axios.delete(`${BACKEND_API_URL}/albums/${albumId}`);
    return response.data;
  } catch (error) {
    console.error('Error deleting album:', error);
    throw error;
  }
};

// Function to create a new track
export const createTrack = async (formData: FormData) => {
  try {
    const response = await axios.post(`${BACKEND_API_URL}/tracks/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  } catch (error) {
    console.error('Error creating track:', error);
    throw error;
  }
};

const ApiService = {};
export default ApiService;
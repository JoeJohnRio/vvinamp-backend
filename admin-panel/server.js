const express = require('express');
const cors = require('cors');
const multer = require('multer');
const path = require('path');
const fs = require('fs');
const { v4: uuidv4 } = require('uuid');
const { Pool } = require('pg');

const app = express();
const PORT = process.env.PORT || 8081;

// Clear problematic environment variables that might interfere
delete process.env.PGHOST;
delete process.env.PGPORT;
delete process.env.PGUSER;
delete process.env.PGDATABASE;
delete process.env.PGPASSWORD;

console.log('Environment variables cleared');

// PostgreSQL connection pool using direct connection string
const pool = new Pool({
  connectionString: 'postgresql://vvinampadmin:password@127.0.0.1:5432/vvinamp',
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

// Middleware
app.use(cors());
app.use(express.json());

// Multer setup for file uploads
const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    const uploadDir = 'uploads/';
    if (!fs.existsSync(uploadDir)) {
      fs.mkdirSync(uploadDir, { recursive: true });
    }
    cb(null, uploadDir);
  },
  filename: (req, file, cb) => {
    cb(null, Date.now() + '-' + file.originalname);
  }
});
const upload = multer({ storage });

// Initialize database tables if they don't exist
async function initializeDatabase() {
  try {
    // Enable UUID extension if not already enabled
    await pool.query('CREATE EXTENSION IF NOT EXISTS "uuid-ossp";');
    
    // Create artists table if it doesn't exist
    await pool.query(`
      CREATE TABLE IF NOT EXISTS artists (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        name VARCHAR(255) NOT NULL,
        bio TEXT,
        profile_image TEXT,
        vertical_video_header_url TEXT,
        verified BOOLEAN DEFAULT FALSE,
        monthly_listeners INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP
      )
    `);
    
    // Create genres table if it doesn't exist
    await pool.query(`
      CREATE TABLE IF NOT EXISTS genres (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        name VARCHAR(255) NOT NULL,
        description TEXT
      )
    `);
    
    // Create albums table if it doesn't exist
    await pool.query(`
      CREATE TABLE IF NOT EXISTS albums (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        title VARCHAR(255) NOT NULL,
        artist_id UUID REFERENCES artists(id),
        release_date DATE,
        cover_art TEXT,
        genre_id UUID REFERENCES genres(id),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `);
    
    // Create tracks table if it doesn't exist
    await pool.query(`
      CREATE TABLE IF NOT EXISTS tracks (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        title VARCHAR(255) NOT NULL,
        duration INTEGER,
        audio_file_url VARCHAR(500),
        stream_count INTEGER DEFAULT 0,
        album_id UUID REFERENCES albums(id),
        explicit BOOLEAN DEFAULT FALSE,
        lyrics TEXT,
        release_date DATE
      )
    `);
    
    // Insert sample data if tables are empty
    const artistResult = await pool.query('SELECT COUNT(*) FROM artists');
    const trackResult = await pool.query('SELECT COUNT(*) FROM tracks');
    
    const artistCount = parseInt(artistResult.rows[0].count);
    const trackCount = parseInt(trackResult.rows[0].count);
    
    if (artistCount === 0) {
      await pool.query(`
        INSERT INTO artists (name, bio, verified, monthly_listeners) VALUES
        ('Sample Artist', 'A sample artist for testing', TRUE, 1000),
        ('Another Artist', 'Another sample artist', FALSE, 500)
      `);
    }
    
    if (trackCount === 0) {
      // Get artist IDs for foreign key references
      const artists = await pool.query('SELECT id FROM artists LIMIT 2');
      
      if (artists.rows.length >= 2) {
        const artist1Id = artists.rows[0].id;
        const artist2Id = artists.rows[1].id;
        
        // Create sample albums
        await pool.query(`
          INSERT INTO albums (title, artist_id, release_date) VALUES
          ('Sample Album 1', $1, CURRENT_DATE),
          ('Sample Album 2', $2, CURRENT_DATE)
        `, [artist1Id, artist2Id]);
        
        // Get album IDs for foreign key references
        const albums = await pool.query('SELECT id FROM albums LIMIT 2');
        
        if (albums.rows.length >= 2) {
          const album1Id = albums.rows[0].id;
          const album2Id = albums.rows[1].id;
          
          // Insert sample tracks
          await pool.query(`
            INSERT INTO tracks (title, duration, audio_file_url, album_id, explicit, lyrics) VALUES
            ('Sample Track 1', 240, '/uploads/sample1.mp3', $1, FALSE, 'These are sample lyrics for track 1'),
            ('Sample Track 2', 180, '/uploads/sample2.mp3', $2, TRUE, 'These are sample lyrics for track 2')
          `, [album1Id, album2Id]);
        }
      }
    }
    
    console.log('Database initialized successfully');
  } catch (err) {
    console.error('Error initializing database:', err);
  }
}

// Initialize the database when the server starts
initializeDatabase();

// API Routes
// Get all tracks
app.get('/api/tracks', async (req, res) => {
  try {
    console.log('Fetching tracks...');

    // Check if the tracks table exists first
    const tableExists = await pool.query(`
      SELECT EXISTS (
        SELECT FROM information_schema.tables
        WHERE table_schema = 'public'
        AND table_name = 'tracks'
      );
    `);

    console.log('Table exists check result:', tableExists.rows[0]);

    if (!tableExists.rows[0].exists) {
      // If table doesn't exist, return empty array
      console.log('Tracks table does not exist');
      return res.json([]);
    }

    // Query with joins to get album and artist information
    console.log('Executing query with joins...');
    const result = await pool.query(`
      SELECT 
        t.id,
        t.title,
        t.duration,
        t.audio_file_url AS "fileName",
        t.release_date AS "uploadDate",
        COALESCE(ar.name, 'Unknown Artist') AS "artist",
        COALESCE(a.title, 'Unknown Album') AS "album",
        t.explicit,
        t.lyrics,
        t.stream_count
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      ORDER BY t.release_date DESC
    `);
    console.log('Query result rows count:', result.rows.length);
    res.json(result.rows);
  } catch (err) {
    console.error('Detailed error fetching tracks:', err.message);
    console.error('Error stack:', err.stack);
    res.status(500).json({ error: 'Internal server error', details: err.message });
  }
});
// Upload a track
app.post('/api/tracks/upload', upload.single('file'), async (req, res) => {
  try {
    const file = req.file;

    if (!file) {
      return res.status(400).json({ error: 'No file uploaded' });
    }

    const { albumId, title } = req.body;
    
    if (!albumId) {
      return res.status(400).json({ error: 'Album ID is required' });
    }

    // --- 1. Get Album & Artist Info for Folder Structure ---
    const albumInfo = await pool.query(`
      SELECT a.title as album_title, ar.name as artist_name
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [albumId]);

    if (albumInfo.rows.length === 0) {
      return res.status(404).json({ error: 'Album not found' });
    }

    const { album_title, artist_name } = albumInfo.rows[0];

    // --- 2. Determine Display Title ---
    // Use provided title OR fallback to original filename (without extension)
    const rawTrackTitle = title || file.originalname.replace(/\.[^/.]+$/, "");

    // --- 3. Generate Sanitized Filename ---
    // Clean helper: "My Song!" -> "my_song"
    const sanitizeName = (name) => name.replace(/[^a-zA-Z0-9\s\-_]/g, '').trim().replace(/\s+/g, '_').toLowerCase();

    const sanitizedTitle = sanitizeName(rawTrackTitle);
    
    // Get extension (.mp3, .wav)
    const fileExtension = path.extname(file.originalname).toLowerCase(); 
    const finalFilename = sanitizedTitle + fileExtension; // "my_song.mp3"

    // --- 4. Setup Directories ---
    const sanitizedArtist = sanitizeName(artist_name || 'unknown_artist');
    const sanitizedAlbum = sanitizeName(album_title || 'unknown_album');

    // Path Structure: Storage/artist/{Artist}/album/{Album}/tracks/
    const storageDir = path.join(__dirname, '..', '..', 'Storage');
    const artistDir = path.join(storageDir, 'artist', sanitizedArtist); // Added 'artist' layer
    const albumDir = path.join(artistDir, 'album', sanitizedAlbum);     // Added 'album' layer
    const trackDir = path.join(albumDir, 'track');

    // Recursive creation handles all parents
    if (!fs.existsSync(trackDir)) {
      fs.mkdirSync(trackDir, { recursive: true });
    }

    // --- 5. Move File to Final Destination ---
    const newFilePath = path.join(trackDir, finalFilename);
    
    // SAFE FIX: Use 'file.path' (absolute path from multer) instead of manual assumption
    fs.renameSync(file.path, newFilePath);

    // --- 6. Insert into DB ---
    const duration = Math.floor(Math.random() * 300) + 60; // Mock duration

    const insertResult = await pool.query(
      `INSERT INTO tracks (title, duration, audio_file_url, album_id, explicit, lyrics)
       VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`,
      [
        rawTrackTitle, // Readable title: "My Song!"
        duration,
        null,          // Placeholder URL
        albumId,
        false, 
        '' 
      ]
    );

    const insertedTrack = insertResult.rows[0];

    // --- 7. Generate URL & Update DB ---
    // URL matches the file on disk: /track/{ID}/my_song.mp3
    const idBasedUrl = `https://cirio.my.id/track/${insertedTrack.id}/${sanitizedTitle}${fileExtension}`;
    
    await pool.query(
      `UPDATE tracks SET audio_file_url = $1 WHERE id = $2`,
      [idBasedUrl, insertedTrack.id]
    );
    
    // Update object for response
    insertedTrack.audio_file_url = idBasedUrl;

    // --- 8. Final Response with Details ---
    const trackWithDetails = await pool.query(`
      SELECT t.*, a.title as album_title, ar.name as artist_name
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE t.id = $1
    `, [insertedTrack.id]);

    res.json(trackWithDetails.rows[0]);

  } catch (err) {
    console.error('Error uploading track:', err);
    // Cleanup temp file on error
    if (req.file && req.file.path && fs.existsSync(req.file.path)) {
      try { fs.unlinkSync(req.file.path); } catch (e) {} 
    }
    res.status(500).json({ error: 'Internal server error' });
  }
});



// Update track file
app.put('/api/tracks/:id/file', upload.single('file'), async (req, res) => {
  try {
    const { id } = req.params;
    const file = req.file;

    if (!file) {
      return res.status(400).json({ error: 'No file uploaded' });
    }

    // Get track to find current file path
    const trackResult = await pool.query('SELECT * FROM tracks WHERE id = $1', [id]);
    
    // Check if track exists
    if (trackResult.rows.length === 0) {
      return res.status(404).json({ error: 'Track not found' });
    }

    const track = trackResult.rows[0];
    
    // 1. Get the title (e.g., "Billie Jean")
    const trackTitle = track.title || 'untitled_track';

    // Get album and artist info to create the proper file path
    const albumInfo = await pool.query(`
      SELECT a.title as album_title, ar.name as artist_name
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [track.album_id]);

    if (albumInfo.rows.length === 0) {
      return res.status(404).json({ error: 'Album not found' });
    }

    const { album_title, artist_name } = albumInfo.rows[0];

    // Sanitize names for file system
    // 2. Create the "name_of_the_song" format
    const sanitizeName = (name) => name.replace(/[^a-zA-Z0-9\s\-_]/g, '').trim().replace(/\s+/g, '_').toLowerCase();
    
    const sanitizedTrackName = sanitizeName(trackTitle); // "Billie Jean" -> "billie_jean"
    const sanitizedArtistName = sanitizeName(artist_name);
    const sanitizedAlbumTitle = sanitizeName(album_title);

    // Create the hierarchical folder structure
    const storageDir = path.join(__dirname, '..', '..', 'Storage');
    const artistDir = path.join(storageDir, 'artist', sanitizedArtistName);
    const albumDir = path.join(artistDir, 'album', sanitizedAlbumTitle);
    const trackDir = path.join(albumDir, 'track');

    // Create directories recursively (simpler one-line approach)
    if (!fs.existsSync(trackDir)) {
      fs.mkdirSync(trackDir, { recursive: true });
    }

    // 3. Construct the new filename with the correct extension
    // We try to get extension from originalname (e.g. song.mp3), fallback to uploaded filename
    const fileExtension = path.extname(file.originalname || file.filename).toLowerCase();
    const newFileName = `${sanitizedTrackName}${fileExtension}`; // e.g., "billie_jean.mp3"
    
    const newFilePath = path.join(trackDir, newFileName);

    // Move the uploaded file from temp 'uploads' folder to final destination
    fs.renameSync(path.join(__dirname, 'uploads', file.filename), newFilePath);

    // Calculate duration (random for now)
    const duration = Math.floor(Math.random() * 300) + 60;

    // 4. Update the URL to match the new filename
    const idBasedUrl = `https://cirio.my.id/track/${id}/${sanitizedTrackName}${fileExtension}`;
    
    const result = await pool.query(
      `UPDATE tracks
       SET audio_file_url = $1, duration = $2
       WHERE id = $3
       RETURNING *`,
      [
        idBasedUrl, 
        duration,
        id
      ]
    );

    // Join with album and artist info for response
    const trackWithDetails = await pool.query(`
      SELECT t.*, a.title as album_title, ar.name as artist_name
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE t.id = $1
    `, [result.rows[0].id]);

    res.json(trackWithDetails.rows[0]);
  } catch (err) {
    console.error('Error updating track file:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Update track metadata
app.put('/api/tracks/:id', async (req, res) => {
  try {
    const { id } = req.params;
    const { title, duration, explicit, lyrics, albumId } = req.body;

    // First, get the current track to check if title is changing and get the current file path
    const currentTrackResult = await pool.query('SELECT * FROM tracks WHERE id = $1', [id]);
    if (currentTrackResult.rows.length === 0) {
      return res.status(404).json({ error: 'Track not found' });
    }
    const currentTrack = currentTrackResult.rows[0];

    // Build dynamic query based on provided fields
    let query = 'UPDATE tracks SET ';
    const values = [];
    const fields = [];

    if (title !== undefined) {
      fields.push(`title = $${fields.length + 1}`);
      values.push(title);
    }
    if (duration !== undefined) {
      fields.push(`duration = $${fields.length + 1}`);
      values.push(duration);
    }
    if (explicit !== undefined) {
      fields.push(`explicit = $${fields.length + 1}`);
      values.push(explicit);
    }
    if (lyrics !== undefined) {
      fields.push(`lyrics = $${fields.length + 1}`);
      values.push(lyrics);
    }
    if (albumId !== undefined) {
      fields.push(`album_id = $${fields.length + 1}`);
      values.push(albumId);
    }

    if (fields.length === 0) {
      return res.status(400).json({ error: 'No fields to update' });
    }

    query += fields.join(', ') + ` WHERE id = $${fields.length + 1} RETURNING *`;
    values.push(id);

    const result = await pool.query(query, values);

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Track not found' });
    }

    // If title was updated, we need to rename the file to match the new title
    if (title !== undefined && title !== currentTrack.title) {
      try {
        // Parse the current audio_file_url to extract the file extension
        const urlParts = currentTrack.audio_file_url.split('/');
        const fileNameWithExt = urlParts[urlParts.length - 1];
        const fileExtension = path.extname(fileNameWithExt);
        
        // Get album and artist info to recreate the file path
        const albumInfo = await pool.query(`
          SELECT a.title as album_title, ar.name as artist_name
          FROM albums a
          LEFT JOIN artists ar ON a.artist_id = ar.id
          WHERE a.id = $1
        `, [result.rows[0].album_id]);

        if (albumInfo.rows.length > 0) {
          const { album_title, artist_name } = albumInfo.rows[0];
          
          // Sanitize names for file system
          const sanitizeName = (name) => name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
          const sanitizedArtist = sanitizeName(artist_name || 'Unknown Artist');
          const sanitizedAlbum = sanitizeName(album_title || 'Unknown Album');

          // Construct the old and new file paths
          const storageDir = path.join(__dirname, '..', '..', 'Storage');
          const artistDir = path.join(storageDir, sanitizedArtist);
          const albumDir = path.join(artistDir, sanitizedAlbum);
          const trackDir = path.join(albumDir, 'track');
          
          const oldFileName = path.parse(fileNameWithExt).name + fileExtension;
          const newFileName = title + fileExtension;
          
          const oldFilePath = path.join(trackDir, oldFileName);
          const newFilePath = path.join(trackDir, newFileName);
          
          // Rename the file if it exists
          if (fs.existsSync(oldFilePath)) {
            fs.renameSync(oldFilePath, newFilePath);
            
            // Update the audio_file_url in the database to reflect the new filename
            const newAudioFileUrl = `https://cirio.my.id/track/${id}/${title}${fileExtension}`;
            await pool.query(
              `UPDATE tracks SET audio_file_url = $1 WHERE id = $2`,
              [newAudioFileUrl, id]
            );
          }
        }
      } catch (renameErr) {
        console.error('Error renaming track file:', renameErr);
        // Don't return an error here as the database update was successful
      }
    }

    // Return track with album and artist info
    const trackWithDetails = await pool.query(`
      SELECT t.*, a.title as album_title, ar.name as artist_name
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE t.id = $1
    `, [result.rows[0].id]);

    res.json(trackWithDetails.rows[0]);
  } catch (err) {
    console.error('Error updating track:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Delete a track
app.delete('/api/tracks/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // Get the track to find its file path
    const trackResult = await pool.query('SELECT audio_file_url FROM tracks WHERE id = $1', [id]);
    
    if (trackResult.rows.length === 0) {
      return res.status(404).json({ error: 'Track not found' });
    }

    const track = trackResult.rows[0];
    
    // Delete the track record
    const result = await pool.query('DELETE FROM tracks WHERE id = $1 RETURNING id', [id]);

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Track not found' });
    }

    // Optionally, delete the physical file (uncomment if you want to delete files)
    /*
    try {
      const filePath = path.join(__dirname, track.audio_file_url);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
      }
    } catch (fileErr) {
      console.error('Error deleting file:', fileErr);
      // Don't return error here as the DB record was deleted successfully
    }
    */

    res.json({ success: true });
  } catch (err) {
    console.error('Error deleting track:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Serve static files from uploads directory
app.use('/uploads', express.static('uploads'));

// Get dashboard statistics
app.get('/api/stats', async (req, res) => {
  try {
    const trackCount = await pool.query('SELECT COUNT(*) FROM tracks');
    const artistCount = await pool.query('SELECT COUNT(*) FROM artists');
    const albumCount = await pool.query('SELECT COUNT(*) FROM albums');

    const stats = {
      totalTracks: parseInt(trackCount.rows[0].count),
      totalArtists: parseInt(artistCount.rows[0].count),
      totalAlbums: parseInt(albumCount.rows[0].count)
    };

    res.json(stats);
  } catch (err) {
    console.error('Error fetching stats:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Get all artists
app.get('/api/artists', async (req, res) => {
  try {
    const result = await pool.query(`
      SELECT 
        id,
        name,
        bio,
        profile_image AS "profileImage",
        verified,
        monthly_listeners AS "monthlyListeners",
        created_at AS "createdAt",
        updated_at AS "updatedAt"
      FROM artists
      ORDER BY name
    `);
    res.json(result.rows);
  } catch (err) {
    console.error('Error fetching artists:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Get all albums
app.get('/api/albums', async (req, res) => {
  try {
    const result = await pool.query(`
      SELECT 
        a.id,
        a.title,
        a.artist_id AS "artistId",
        ar.name AS "artistName",
        a.release_date AS "releaseDate",
        a.cover_art AS "coverArt"
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      ORDER BY a.title
    `);
    res.json(result.rows);
  } catch (err) {
    console.error('Error fetching albums:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Multer setup for artist profile image uploads
const artistImageStorage = multer.diskStorage({
  destination: (req, file, cb) => {
    // Use a temporary location for all requests
    // We'll move the file to the correct location in the route handlers
    const storageDir = path.join(__dirname, '..', '..', 'Storage', 'temp');
    if (!fs.existsSync(storageDir)) {
      fs.mkdirSync(storageDir, { recursive: true });
    }
    cb(null, storageDir);
  },
  filename: (req, file, cb) => {
    // Get artist name from request body
    let artistName = req.body.name || 'unknown_artist';
    
    // For update requests, we'll handle the filename differently if needed
    // For now, use the name from the request body if provided
    const sanitizedArtistName = artistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedFileName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
    cb(null, sanitizedFileName + '.jpg');
  }
});
const uploadArtistImage = multer({ storage: artistImageStorage });

// Create a new artist
app.post('/api/artists', uploadArtistImage.single('profileImage'), async (req, res) => {
  try {
    const { name, bio, verified = false } = req.body;
    
    // Determine the profile image path
    let profileImagePath = null;
    
    if (req.file) {
      // Generate readable folder name for file storage
      const sanitizedArtistName = name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedFolderNameClean = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');

      // Create the proper directory structure: Storage/artist/artist_name/
      const artistDir = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedFolderNameClean);
      if (!fs.existsSync(artistDir)) {
        fs.mkdirSync(artistDir, { recursive: true });
      }

      // Define the new file path
      const newFilePath = path.join(artistDir, sanitizedFolderNameClean + '.jpg');

      // Move the uploaded file to the correct location
      fs.renameSync(req.file.path, newFilePath);
    }

    // Insert the artist record first to get the ID
    const insertResult = await pool.query(
      `INSERT INTO artists (name, bio, profile_image, verified, monthly_listeners)
       VALUES ($1, $2, $3, $4, $5) RETURNING *`,
      [name, bio || null, null, verified, 0]  // Initially set profile_image to null
    );
    
    const insertedArtist = insertResult.rows[0];
    
    // If there's a profile image, update the record with the ID-based URL
    if (req.file) {
      // Generate ID-based CDN URL for the image
      const sanitizedArtistName = name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedFolderNameClean = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
      const idBasedUrl = `https://cirio.my.id/artist/${insertedArtist.id}/${sanitizedFolderNameClean}.jpg`;
      
      // Update the record with the ID-based URL
      await pool.query(
        `UPDATE artists SET profile_image = $1 WHERE id = $2`,
        [idBasedUrl, insertedArtist.id]
      );
      
      // Update the insertedArtist object to return the correct URL
      insertedArtist.profile_image = idBasedUrl;
    }

    res.json(insertedArtist);
  } catch (err) {
    console.error('Error creating artist:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Multer setup for album cover art uploads during creation
const albumCoverCreateStorage = multer.diskStorage({
  destination: (req, file, cb) => {
    // Use a temporary location for all requests
    // We'll move the file to the correct location in the route handlers
    const storageDir = path.join(__dirname, '..', '..', 'Storage', 'temp');
    if (!fs.existsSync(storageDir)) {
      fs.mkdirSync(storageDir, { recursive: true });
    }
    cb(null, storageDir);
  },
  filename: (req, file, cb) => {
    // Generate a unique filename for the cover art based on album title
    const sanitizedAlbumTitle = req.body.title ? req.body.title.replace(/[^a-zA-Z0-9\s\-_]/g, '_') : 'album_cover';
    const sanitizedFileName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
    const ext = path.extname(file.originalname);
    cb(null, sanitizedFileName + ext);
  }
});
const uploadAlbumCoverCreate = multer({ storage: albumCoverCreateStorage });

// Create a new album
app.post('/api/albums', uploadAlbumCoverCreate.single('coverArt'), async (req, res) => {
  try {
    let coverArtUrl = null;
    
    // Insert the album record first to get the ID (without cover art initially)
    const insertResult = await pool.query(
      `INSERT INTO albums (title, artist_id, release_date, cover_art)
       VALUES ($1, $2, $3, $4) RETURNING *`,
      [req.body.title, req.body.artistId, req.body.releaseDate || null, null]  // Initially set cover_art to null
    );
    
    const insertedAlbum = insertResult.rows[0];
    
    if (req.file) {
      // Get artist name to create the proper folder structure
      let artistName = '';
      if (req.body.artistId) {
        const artistResult = await pool.query('SELECT name FROM artists WHERE id = $1', [req.body.artistId]);
        if (artistResult.rows.length > 0) {
          artistName = artistResult.rows[0].name;
        }
      }

      // Generate readable folder names for file storage
      const sanitizedArtistName = artistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedArtistFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
      const sanitizedAlbumTitle = req.body.title ? req.body.title.replace(/[^a-zA-Z0-9\s\-_]/g, '_') : 'album_cover';
      const sanitizedAlbumFolderName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
      
      // Create the proper directory structure: Storage/artist/artist_name/album/album_name/
      const artistDir = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedArtistFolderName);
      const albumDir = path.join(artistDir, 'album', sanitizedAlbumFolderName);
      if (!fs.existsSync(artistDir)) {
        fs.mkdirSync(artistDir, { recursive: true });
      }
      if (!fs.existsSync(albumDir)) {
        fs.mkdirSync(albumDir, { recursive: true });
      }

      // Define the new file path
      const fileNameWithoutExt = path.parse(req.file.filename).name;
      const newFilePath = path.join(albumDir, fileNameWithoutExt + '.jpg');
      await fs.promises.copyFile(req.file.path, newFilePath);
      await fs.promises.unlink(req.file.path); // Remove the temp file
      
      // Generate ID-based CDN URL for the cover art
      coverArtUrl = `https://cirio.my.id/album/${insertedAlbum.id}/${fileNameWithoutExt}.jpg`;
      
      // Update the album record with the ID-based URL
      await pool.query(
        `UPDATE albums SET cover_art = $1 WHERE id = $2`,
        [coverArtUrl, insertedAlbum.id]
      );
      
      // Update the insertedAlbum object to return the correct URL
      insertedAlbum.cover_art = coverArtUrl;
    }

    // If coverArt is provided in the body (not as a file upload), use that instead
    if (req.body.coverArt && !req.file) {
      coverArtUrl = req.body.coverArt;
      // Update the album record with the provided URL
      await pool.query(
        `UPDATE albums SET cover_art = $1 WHERE id = $2`,
        [coverArtUrl, insertedAlbum.id]
      );
    }

    // Also return artist name with the album
    const albumWithArtist = await pool.query(`
      SELECT 
        a.id,
        a.title,
        a.artist_id AS "artistId",
        ar.name AS "artistName",
        a.release_date AS "releaseDate",
        a.cover_art AS "coverArt"
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [insertedAlbum.id]);

    res.json(albumWithArtist.rows[0]);
  } catch (err) {
    console.error('Error creating album:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Update an artist
app.put('/api/artists/:id', uploadArtistImage.single('profileImage'), async (req, res) => {
  try {
    const { id } = req.params;
    const { name, bio, verified } = req.body;

    // Get the current artist to get the name if not provided in update
    const currentArtistResult = await pool.query('SELECT name FROM artists WHERE id = $1', [id]);
    if (currentArtistResult.rows.length === 0) {
      return res.status(404).json({ error: 'Artist not found' });
    }
    const currentArtistName = currentArtistResult.rows[0].name;

    // Determine the profile image path
    let profileImagePath = null;
    
    if (req.file) {
      // Generate readable folder name for file storage
      const sanitizedNewArtistName = (name || currentArtistName).replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedNewFolderName = sanitizedNewArtistName.toLowerCase().replace(/\s+/g, '_');

      // Create the proper directory structure: Storage/artist/artist_name/
      const newArtistDir = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedNewFolderName);
      if (!fs.existsSync(newArtistDir)) {
        fs.mkdirSync(newArtistDir, { recursive: true });
      }

      // Define the new file path
      const newFilePath = path.join(newArtistDir, sanitizedNewFolderName + '.jpg');
      await fs.promises.copyFile(req.file.path, newFilePath);
      await fs.promises.unlink(req.file.path);

      // Generate ID-based CDN URL for the image
      profileImagePath = `https://cirio.my.id/artist/${id}/${sanitizedNewFolderName}.jpg`;
    }

    let query;
    let queryParams;

    if (profileImagePath) {
      // Update with new profile image
      query = `UPDATE artists
               SET name = $1, bio = $2, profile_image = $3, verified = $4, updated_at = CURRENT_TIMESTAMP
               WHERE id = $5
               RETURNING *`;
      queryParams = [name, bio || null, profileImagePath, verified, id];
    } else {
      // Update without changing profile image
      query = `UPDATE artists
               SET name = $1, bio = $2, verified = $3, updated_at = CURRENT_TIMESTAMP
               WHERE id = $4
               RETURNING *`;
      queryParams = [name, bio || null, verified, id];
    }

    const result = await pool.query(query, queryParams);

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Artist not found' });
    }

    res.json(result.rows[0]);
  } catch (err) {
    console.error('Error updating artist:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Delete an artist
app.delete('/api/artists/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // First, delete all albums by this artist
    await pool.query('DELETE FROM albums WHERE artist_id = $1', [id]);

    // Then delete all tracks from those albums (this might be handled by foreign key constraints)
    // Or we can delete tracks that belong to albums of this artist
    // Note: This assumes cascading deletes or manual cleanup
    
    // Finally, delete the artist
    const result = await pool.query('DELETE FROM artists WHERE id = $1 RETURNING id', [id]);

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Artist not found' });
    }

    res.json({ success: true });
  } catch (err) {
    console.error('Error deleting artist:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Multer setup for album cover art uploads
const albumCoverStorage = multer.diskStorage({
  destination: (req, file, cb) => {
    // Use a temporary location for all requests
    // We'll move the file to the correct location in the route handlers
    const storageDir = path.join(__dirname, '..', '..', 'Storage', 'temp');
    if (!fs.existsSync(storageDir)) {
      fs.mkdirSync(storageDir, { recursive: true });
    }
    cb(null, storageDir);
  },
  filename: (req, file, cb) => {
    // Generate a unique filename for the cover art based on album title
    const sanitizedAlbumTitle = req.body.title ? req.body.title.replace(/[^a-zA-Z0-9\s\-_]/g, '_') : 'album_cover';
    const sanitizedFileName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
    const ext = path.extname(file.originalname);
    cb(null, sanitizedFileName + ext);
  }
});
const uploadAlbumCover = multer({ storage: albumCoverStorage });

// Update an album
app.put('/api/albums/:id', uploadAlbumCover.single('coverArt'), async (req, res) => {
  try {
    const { id } = req.params;
    
    // First, get the current album information to create the proper folder structure
    const currentAlbumResult = await pool.query(`
      SELECT a.title, ar.name as artist_name
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [id]);
    
    if (currentAlbumResult.rows.length === 0) {
      return res.status(404).json({ error: 'Album not found' });
    }
    
    const currentAlbum = currentAlbumResult.rows[0];
    const artistName = currentAlbum.artist_name || 'unknown_artist';
    
    // Handle the case where there might be file upload
    let coverArtUrl = null;
    
    if (req.file) {
      // Generate readable folder names for file storage
      const sanitizedArtistName = artistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedArtistFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
      const sanitizedAlbumTitle = req.body.title ? req.body.title.replace(/[^a-zA-Z0-9\s\-_]/g, '_') : currentAlbum.title.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedAlbumFolderName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
      
      // Create the proper directory structure: Storage/artist/artist_name/album/album_name/
      const artistDir = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedArtistFolderName);
      const albumDir = path.join(artistDir, 'album', sanitizedAlbumFolderName);
      if (!fs.existsSync(artistDir)) {
        fs.mkdirSync(artistDir, { recursive: true });
      }
      if (!fs.existsSync(albumDir)) {
        fs.mkdirSync(albumDir, { recursive: true });
      }

      // Define the new file path
      const fileNameWithoutExt = path.parse(req.file.filename).name;
      const newFilePath = path.join(albumDir, fileNameWithoutExt + '.jpg');
      await fs.promises.copyFile(req.file.path, newFilePath);
      await fs.promises.unlink(req.file.path); // Remove the temp file
      
      // Generate ID-based CDN URL for the cover art
      coverArtUrl = `https://cirio.my.id/album/${id}/${fileNameWithoutExt}.jpg`;
    }

    // Get the other fields from the request body
    const { title, artistId, releaseDate } = req.body;

    // If coverArt is provided in the body (not as a file upload), use that instead
    if (req.body.coverArt && !req.file) {
      coverArtUrl = req.body.coverArt;
    }

    const result = await pool.query(
      `UPDATE albums
       SET title = $1, artist_id = $2, release_date = $3, cover_art = $4
       WHERE id = $5
       RETURNING *`,
      [title, artistId, releaseDate || null, coverArtUrl || null, id]
    );

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Album not found' });
    }

    // Return album with artist name
    const albumWithArtist = await pool.query(`
      SELECT 
        a.id,
        a.title,
        a.artist_id AS "artistId",
        ar.name AS "artistName",
        a.release_date AS "releaseDate",
        a.cover_art AS "coverArt"
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [result.rows[0].id]);

    res.json(albumWithArtist.rows[0]);
  } catch (err) {
    console.error('Error updating album:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Delete an album
app.delete('/api/albums/:id', async (req, res) => {
  try {
    const { id } = req.params;

    // Delete all tracks in this album first
    await pool.query('DELETE FROM tracks WHERE album_id = $1', [id]);

    // Then delete the album
    const result = await pool.query('DELETE FROM albums WHERE id = $1 RETURNING id', [id]);

    if (result.rowCount === 0) {
      return res.status(404).json({ error: 'Album not found' });
    }

    res.json({ success: true });
  } catch (err) {
    console.error('Error deleting album:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Basic route for health check
app.get('/', (req, res) => {
  res.json({ message: 'Admin Panel API Server is running!' });
});

// Test route
app.get('/test', async (req, res) => {
  try {
    console.log('Test route called');
    res.json({ message: 'Test route works!' });
  } catch (err) {
    console.error('Error in test route:', err);
    res.status(500).json({ error: 'Test route error', details: err.message });
  }
});

// Mapping endpoints to serve files by ID
// Serve artist profile image by ID
app.get('/artist/:artistId/:fileName', async (req, res) => {
  try {
    const { artistId, fileName } = req.params;
    
    // Look up the artist by ID to get the actual file path
    const artistResult = await pool.query(
      'SELECT name FROM artists WHERE id = $1', 
      [artistId]
    );
    
    if (artistResult.rows.length === 0) {
      return res.status(404).send('Artist not found');
    }
    
    const artistName = artistResult.rows[0].name;
    const sanitizedArtistName = artistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
    
    // Construct the actual file path
    const filePath = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedFolderName, fileName);
    // Check if file exists before sending
    if (!fs.existsSync(filePath)) {
      return res.status(404).send('File not found');
    }
    
    res.sendFile(filePath);
  } catch (err) {
    console.error('Error serving artist file:', err);
    res.status(500).send('Internal server error');
  }
});

// Serve album cover by ID
app.get('/album/:albumId/:fileName', async (req, res) => {
  try {
    const { albumId, fileName } = req.params;
    
    // Look up the album by ID to get the artist and album names
    const albumResult = await pool.query(`
      SELECT a.title as album_title, ar.name as artist_name
      FROM albums a
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE a.id = $1
    `, [albumId]);
    
    if (albumResult.rows.length === 0) {
      return res.status(404).send('Album not found');
    }
    
    const { album_title, artist_name } = albumResult.rows[0];
    
    const sanitizedArtistName = artist_name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedArtistFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
    const sanitizedAlbumTitle = album_title.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedAlbumFolderName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
    
    // Construct the actual file path
    const filePath = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedArtistFolderName, 'album', sanitizedAlbumFolderName, fileName);
    
    // Check if file exists before sending
    if (!fs.existsSync(filePath)) {
      return res.status(404).send('File not found');
    }
    
    res.sendFile(filePath);
  } catch (err) {
    console.error('Error serving album file:', err);
    res.status(500).send('Internal server error');
  }
});

// Serve track file by ID
app.get('/track/:trackId/:fileName', async (req, res) => {
  try {
    const { trackId, fileName } = req.params;
    
    // Look up the track by ID to get the album and artist information
    const trackResult = await pool.query(`
      SELECT a.title as album_title, ar.name as artist_name
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE t.id = $1
    `, [trackId]);
    
    if (trackResult.rows.length === 0) {
      return res.status(404).send('Track not found');
    }
    
    const { album_title, artist_name } = trackResult.rows[0];
    
    const sanitizedArtistName = artist_name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedArtistFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
    const sanitizedAlbumTitle = album_title.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedAlbumFolderName = sanitizedAlbumTitle.toLowerCase().replace(/\s+/g, '_');
    
    // Construct the actual file path
    const filePath = path.join(__dirname, '..', '..', 'Storage', 'artist', sanitizedArtistFolderName, 'album', sanitizedAlbumFolderName, 'tracks', fileName);
    
    // Check if file exists before sending
    if (!fs.existsSync(filePath)) {
      return res.status(404).send('File not found');
    }
    
    res.sendFile(filePath);
  } catch (err) {
    console.error('Error serving track file:', err);
    res.status(500).send('Internal server error');
  }
});

app.listen(PORT, () => {
  console.log(`Admin Panel API Server is running on port ${PORT}`);
});
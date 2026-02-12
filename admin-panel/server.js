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

    // Get album ID from request body
    const { albumId, title } = req.body;
    
    if (!albumId) {
      return res.status(400).json({ error: 'Album ID is required' });
    }

    // Get album and artist info to create the proper file path
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
    
    // Sanitize names for file system
    const sanitizeName = (name) => name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
    const sanitizedArtist = sanitizeName(artist_name || 'Unknown Artist');
    const sanitizedAlbum = sanitizeName(album_title || 'Unknown Album');
    
    // Create the hierarchical folder structure in the Storage directory
    const storageDir = path.join(__dirname, '..', '..', 'Storage');
    const artistDir = path.join(storageDir, sanitizedArtist);
    const albumDir = path.join(artistDir, sanitizedAlbum);
    
    // Create directories if they don't exist
    if (!fs.existsSync(storageDir)) {
      fs.mkdirSync(storageDir, { recursive: true });
    }
    if (!fs.existsSync(artistDir)) {
      fs.mkdirSync(artistDir, { recursive: true });
    }
    if (!fs.existsSync(albumDir)) {
      fs.mkdirSync(albumDir, { recursive: true });
    }
    
    // Move the uploaded file to the proper location
    const newFilePath = path.join(albumDir, file.filename);
    fs.renameSync(path.join(__dirname, 'uploads', file.filename), newFilePath);
    
    // Calculate duration (in a real implementation, you'd use a library like ffmpeg to get the actual duration)
    // For now, we'll keep the random duration approach
    const duration = Math.floor(Math.random() * 300) + 60;

    // Insert the track record with the proper file path
    const result = await pool.query(
      `INSERT INTO tracks (title, duration, audio_file_url, album_id, explicit, lyrics)
       VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`,
      [
        title || file.originalname.replace(/\.[^/.]+$/, ""), // Use provided title or derive from filename
        duration,
        path.relative(__dirname, newFilePath), // Store relative path to the file
        albumId,
        false, // explicit
        '' // lyrics
      ]
    );

    // Also join with album and artist info for the response
    const trackWithDetails = await pool.query(`
      SELECT t.*, a.title as album_title, ar.name as artist_name
      FROM tracks t
      LEFT JOIN albums a ON t.album_id = a.id
      LEFT JOIN artists ar ON a.artist_id = ar.id
      WHERE t.id = $1
    `, [result.rows[0].id]);

    res.json(trackWithDetails.rows[0]);
  } catch (err) {
    console.error('Error uploading track:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Update track metadata
app.put('/api/tracks/:id', async (req, res) => {
  try {
    const { id } = req.params;
    const { title, duration, explicit, lyrics, albumId } = req.body;

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
      // Generate CDN URL for the image
      const sanitizedArtistName = name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
      profileImagePath = `https://cirio.my.id/artist/${sanitizedFolderName}/${sanitizedFolderName}.jpg`;
      
      // Move the uploaded file to the correct location
      const sanitizedArtistNameClean = name.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedFolderNameClean = sanitizedArtistNameClean.toLowerCase().replace(/\s+/g, '_');
      
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

    const result = await pool.query(
      `INSERT INTO artists (name, bio, profile_image, verified, monthly_listeners)
       VALUES ($1, $2, $3, $4, $5) RETURNING *`,
      [name, bio || null, profileImagePath, verified, 0]
    );

    res.json(result.rows[0]);
  } catch (err) {
    console.error('Error creating artist:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Create a new album
app.post('/api/albums', async (req, res) => {
  try {
    const { title, artistId, releaseDate, coverArt } = req.body;

    const result = await pool.query(
      `INSERT INTO albums (title, artist_id, release_date, cover_art)
       VALUES ($1, $2, $3, $4) RETURNING *`,
      [title, artistId, releaseDate || null, coverArt || null]
    );

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
    `, [result.rows[0].id]);

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
      // Generate CDN URL for the image
      const sanitizedArtistName = (name || currentArtistName).replace(/[^a-zA-Z0-9\s\-_]/g, '_');
      const sanitizedFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');
      profileImagePath = `https://cirio.my.id/artist/${sanitizedFolderName}/${sanitizedFolderName}.jpg`;
      
      // Move the uploaded file to the correct location
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

// Update an album
app.put('/api/albums/:id', async (req, res) => {
  try {
    const { id } = req.params;
    const { title, artistId, releaseDate, coverArt } = req.body;

    const result = await pool.query(
      `UPDATE albums 
       SET title = $1, artist_id = $2, release_date = $3, cover_art = $4
       WHERE id = $5
       RETURNING *`,
      [title, artistId, releaseDate || null, coverArt || null, id]
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

app.listen(PORT, () => {
  console.log(`Admin Panel API Server is running on port ${PORT}`);
});
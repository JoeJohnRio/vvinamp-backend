const { Pool } = require('pg');

// Clear problematic environment variables that might interfere
delete process.env.PGHOST;
delete process.env.PGPORT;
delete process.env.PGUSER;
delete process.env.PGDATABASE;
delete process.env.PGPASSWORD;

// PostgreSQL connection pool using direct connection string
const pool = new Pool({
  connectionString: 'postgresql://vvinampadmin:password@127.0.0.1:5432/vvinamp',
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

async function testConnection() {
  try {
    console.log('Testing database connection...');
    
    // Test basic connection
    const result = await pool.query('SELECT NOW()');
    console.log('Basic connection successful:', result.rows[0]);
    
    // Check if tracks table exists
    const tableCheck = await pool.query(`
      SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'tracks'
      );
    `);
    console.log('Tracks table exists:', tableCheck.rows[0].exists);
    
    // If table exists, try to query it
    if (tableCheck.rows[0].exists) {
      const tracks = await pool.query('SELECT * FROM tracks LIMIT 5');
      console.log('Sample tracks:', tracks.rows);
    } else {
      console.log('Tracks table does not exist yet');
    }
    
    // Check all tables
    const allTables = await pool.query(`
      SELECT table_name 
      FROM information_schema.tables 
      WHERE table_schema = 'public'
    `);
    console.log('All tables in database:', allTables.rows.map(row => row.table_name));
    
  } catch (err) {
    console.error('Database test failed:', err.message);
    console.error('Stack trace:', err.stack);
  } finally {
    await pool.end();
  }
}

testConnection();
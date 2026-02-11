const fs = require('fs');
const path = require('path');

// Test to verify the directory structure
const artistName = "Test Artist";
const sanitizedArtistName = artistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
const sanitizedFolderName = sanitizedArtistName.toLowerCase().replace(/\s+/g, '_');

console.log('Artist name:', artistName);
console.log('Sanitized artist name:', sanitizedArtistName);
console.log('Sanitized folder name:', sanitizedFolderName);

// Check the directory structure that should be created
const expectedDir = path.join(__dirname, '..', '..', 'Storage', 'vvinampStorage', 'artist', sanitizedFolderName);
console.log('Expected directory path:', expectedDir);

// Check if the directory exists
console.log('Directory exists:', fs.existsSync(expectedDir));

// Also test with a different artist name
const anotherArtistName = "Another Test Artist!";
const sanitizedAnotherArtistName = anotherArtistName.replace(/[^a-zA-Z0-9\s\-_]/g, '_');
const sanitizedAnotherFolderName = sanitizedAnotherArtistName.toLowerCase().replace(/\s+/g, '_');

console.log('\nAnother artist name:', anotherArtistName);
console.log('Sanitized another artist name:', sanitizedAnotherArtistName);
console.log('Sanitized another folder name:', sanitizedAnotherFolderName);

const expectedDir2 = path.join(__dirname, '..', '..', 'Storage', 'vvinampStorage', 'artist', sanitizedAnotherFolderName);
console.log('Expected directory path for second artist:', expectedDir2);
console.log('Directory exists:', fs.existsSync(expectedDir2));
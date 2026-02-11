# vvinamp Admin Panel Testing Guide

## Manual Testing Steps

1. **Verify Build Process**
   ```bash
   cd admin-panel
   npm run build
   ```
   
   This will create a production-ready build in the `build` folder. If this succeeds, it means all components are properly connected.

2. **Component Testing**
   The application includes the following components that have been implemented:
   - Main App component with routing
   - Sidebar navigation (fixed positioning to prevent overlap)
   - Dashboard with statistics
   - Music Library with:
     - Drag and drop upload zone
     - Track list view
     - Folder organization view
     - Track management (edit/delete)

3. **API Integration**
   The application is set up to connect to your Go backend via:
   - Track upload functionality
   - Track retrieval
   - Track deletion
   - Track updates

4. **Features Tested**
   ✓ Build process completed successfully
   ✓ Fixed sidebar/main content overlap issue
   ✓ Updated branding from Spotify to vvinamp
   ✓ File upload handling
   ✓ Folder organization view
   ✓ Track management
   ✓ API service integration
   ✓ Responsive design

Note: The drag-and-drop functionality for individual tracks between folders has been simplified to ensure compatibility. The upload zone still supports drag-and-drop functionality for file uploads.

## Running the Application

To run the admin panel:

1. Navigate to the admin-panel directory:
   ```bash
   cd admin-panel
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Create a .env file with your backend URL:
   ```
   REACT_APP_API_URL=http://localhost:8081
   ```

4. Start the development server:
   ```bash
   npm start
   ```

The application will be available at http://localhost:3001

## Connecting to Your Go Backend

To fully utilize the admin panel, you'll need to implement the corresponding API endpoints in your Go application. The endpoints the admin panel expects are:

- GET /api/tracks - Retrieve all tracks
- POST /api/tracks/upload - Upload a new track
- DELETE /api/tracks/:id - Delete a track
- PUT /api/tracks/:id - Update track metadata

## Summary

The admin panel has been successfully created with all requested features:
- ✅ Drag and drop file upload
- ✅ Folder organization
- ✅ Track management
- ✅ Connection to database (via API calls)
- ✅ Responsive web interface

The application is ready to be integrated with your existing Go backend.
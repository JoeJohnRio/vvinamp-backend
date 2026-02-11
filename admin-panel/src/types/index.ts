export interface Track {
  id: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  fileName: string;
  uploadDate: string; // Changed to string since localStorage stores dates as ISO strings
  fileSize?: number;
  fileType?: string;
  explicit?: boolean;
  lyrics?: string;
}

export interface Artist {
  id: string;
  name: string;
  bio?: string;
  profileImage?: string;
  verified?: boolean;
  monthlyListeners?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface Album {
  id: string;
  title: string;
  artistId: string;
  artistName: string;
  releaseDate?: string;
  coverArt?: string;
  createdAt?: string;
}
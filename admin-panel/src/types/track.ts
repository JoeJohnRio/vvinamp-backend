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
}
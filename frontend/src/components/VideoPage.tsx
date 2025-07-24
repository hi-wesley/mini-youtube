import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import api from '../api/axios';

import VideoPlayer from './VideoPlayer';
import CommentArea from './CommentArea';

interface Video {
  ID: string;
  Title: string;
  ObjectName: string;
}

export default function VideoPage() {
  const { id } = useParams();
  const { data: video, isLoading, error } = useQuery<Video>({ 
    queryKey: ['video', id], 
    queryFn: () => api.get(`/v1/videos/${id}`).then(res => res.data),
    enabled: !!id,
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>An error occurred: {error.message}</div>;
  if (!video) return <div>Video not found</div>;

  const videoSrc = `${import.meta.env.VITE_GCS_URL}/${video.ObjectName}`;
  console.log("Video source URL:", videoSrc);

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-end mb-4">
        <Link to="/" className="text-blue-600">Go to Main Page</Link>
      </div>
      <div style={{ display: 'flex', flexDirection: 'row' }}>
        <div style={{ flex: 3 }}>
          <VideoPlayer src={videoSrc} />
        </div>
        <div style={{ flex: 1 }}>
          <CommentArea videoId={video.ID} />
        </div>
      </div>
    </div>
  );
}

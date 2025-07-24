import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import api from '../api/axios';

interface Video {
  ID: string;
  Title: string;
  // Add other video properties here as needed
}

export default function VideoList() {
  const { data: videos, isLoading, error } = useQuery<Video[]>({ 
    queryKey: ['videos'], 
    queryFn: () => api.get('/v1/videos').then(res => res.data)
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>An error occurred: {error.message}</div>;

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Uploaded Videos</h1>
        <Link to="/upload" className="bg-blue-600 text-white px-4 py-2 rounded">
          Upload Video
        </Link>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {videos?.map(video => (
          <Link to={`/watch/${video.ID}`} key={video.ID} className="border rounded-lg overflow-hidden">
            {/* You can add a thumbnail image here later */}
            <div className="p-4">
              <h3 className="font-bold">{video.Title}</h3>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}

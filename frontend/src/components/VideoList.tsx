import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import api from '../api/axios';

// Helper function to format time since video upload
const timeAgo = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  let interval = seconds / 31536000;
  if (interval > 1) {
    return Math.floor(interval) + " years ago";
  }
  interval = seconds / 2592000;
  if (interval > 1) {
    return Math.floor(interval) + " months ago";
  }
  interval = seconds / 86400;
  if (interval > 1) {
    return Math.floor(interval) + " days ago";
  }
  interval = seconds / 3600;
  if (interval > 1) {
    return Math.floor(interval) + " hours ago";
  }
  interval = seconds / 60;
  if (interval > 1) {
    return Math.floor(interval) + " minutes ago";
  }
  return "just now";
};

// Helper function to format view counts
const formatViews = (views: number) => {
  if (views >= 1000000) {
    return `${(views / 1000000).toFixed(1)}M`;
  }
  if (views >= 1000) {
    return `${Math.floor(views / 1000)}K`;
  }
  return views;
};

interface User {
  Username: string;
}

interface Video {
  ID: string;
  Title: string;
  ThumbnailURL: string;
  User: User;
  Views: number;
  CreatedAt: string;
}

export default function VideoList() {
  const { data: videos, isLoading, error } = useQuery<Video[]>({
    queryKey: ['videos'],
    queryFn: () => api.get('/v1/videos').then(res => res.data || []),
  });

  if (isLoading) {
    return <div className="text-center p-10">Loading videos...</div>;
  }

  if (error) {
    return <div className="text-center p-10 text-red-500">An error occurred: {error.message}</div>;
  }

  return (
    <main className="flex-1 p-4 sm:p-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-x-4 gap-y-8">
        {videos?.map(video => (
          <Link to={`/watch/${video.ID}`} key={video.ID} className="flex flex-col">
            <div className="relative">
              <img src={video.ThumbnailURL} alt={video.Title} className="w-full h-auto rounded-xl object-cover aspect-video" />
              {/* Duration can be added here if available in the model */}
            </div>
            <div className="flex items-start mt-3">
              {/* Avatar removed as per request */}
              <div className="ml-3">
                <p className="text-base font-medium text-f1f1f1 leading-tight break-words">{video.Title}</p>
                <p className="text-sm text-zinc-400 mt-1">{video.User.Username}</p>
                <p className="text-sm text-zinc-400">
                  {formatViews(video.Views)} views &bull; {timeAgo(video.CreatedAt)}
                </p>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </main>
  );
}

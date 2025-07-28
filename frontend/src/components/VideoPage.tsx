import { Link, useParams, useNavigate, useLocation } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../api/axios';
import { useContext, useEffect, useRef } from 'react';

import VideoPlayer from './VideoPlayer';
import CommentArea from './CommentArea';
import { AuthCtx } from './AuthProvider';

interface Video {
  ID: string;
  Title: string;
  Description: string;
  ObjectName: string;
  User: {
    Username: string;
  };
  CreatedAt: string;
  Views: number;
  Summary: string;
  SummaryModel: string;
  Likes: number;
  IsLiked: boolean;
}

export default function VideoPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const viewIncremented = useRef(false);
  const auth = useContext(AuthCtx);

  const { data: video, isLoading, error } = useQuery<Video>({
    queryKey: ['video', id],
    queryFn: () => api.get(`/v1/videos/${id}`).then(res => res.data),
    enabled: !!id && !auth?.loading, // <-- Wait for auth to be loaded
  });

  const likeMutation = useMutation({
    mutationFn: () => api.post(`/v1/videos/${id}/like`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['video', id] });
    },
  });

  const handleLike = () => {
    if (!auth?.user) {
      alert('You are not logged in');
      return;
    }
    likeMutation.mutate();
  };

  useEffect(() => {
    if (id && !viewIncremented.current) {
      const incrementView = async () => {
        try {
          await api.post(`/v1/videos/${id}/view`);
          queryClient.invalidateQueries({ queryKey: ['video', id] });
        } catch (err) {
          console.error("Failed to increment view count", err);
        }
      };
      incrementView();
      viewIncremented.current = true;
    }
  }, [id, queryClient]);

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>An error occurred: {error.message}</div>;
  if (!video) return <div>Video not found</div>;

  const videoSrc = `${import.meta.env.VITE_GCS_URL}/${video.ObjectName}`;

  return (
    <div className="container mx-auto p-4 max-w-4xl w-full">
      
      <div className="flex flex-col gap-4">
        <div>
          <VideoPlayer src={videoSrc} autoPlay />
          <div className="mt-4">
            <h1 className="text-2xl font-bold">{video.Title}</h1>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }} className="mt-2">
              <div className="text-gray-600">Uploaded by {video.User.Username}</div>
              <div className="like-wrapper">
                <div className="container" onClick={handleLike}>
                  <input 
                    type="checkbox" 
                    checked={video.IsLiked && !!auth?.user} 
                    onChange={handleLike}
                    readOnly
                  />
                  <svg viewBox="0 0 24 24">
                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                  </svg>
                </div>
              </div>
            </div>

            <style jsx>{`
              .like-wrapper {
                position: relative;
                display: flex;
                align-items: center;
                text-align: center;
                justify-content: center;
                cursor: pointer;
              }
              
              .like-wrapper::before {
                content: '';
                position: absolute;
                inset: -1px;
                background: linear-gradient(to right, #93c5fd, #3b82f6);
                border-radius: 9999px;
                opacity: 0;
                transition: opacity 0.3s;
                filter: blur(1px);
                z-index: 0;
              }
              
              .like-wrapper:hover::before {
                opacity: 1;
              }
              
              .like-wrapper > .container {
                position: relative;
                background: white;
                border: 1px solid #e5e7eb;
                border-radius: 9999px;
                padding: 0.5rem;
                transition: background-color 0.3s;
                z-index: 1;
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
                user-select: none;
              }
              
              .like-wrapper:hover > .container {
                background: #f9fafb;
              }

              .container input {
                position: absolute;
                opacity: 0;
                cursor: pointer;
                height: 0;
                width: 0;
              }

              .container svg {
                position: relative;
                top: 0;
                left: 0;
                height: 20px;
                width: 20px;
                transition: all 0.3s;
                fill: #666;
              }

              .container svg:hover {
                transform: scale(1.1) rotate(-10deg);
              }

              .container input:checked ~ svg {
                fill: #2196F3;
              }
            `}</style>
            <div className="mt-4 p-4 bg-gray-100 rounded-lg">
              <p className="text-sm font-medium text-gray-700 mb-1">{video.Views.toLocaleString()} views • Uploaded {new Date(video.CreatedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</p>
              <p className="text-base whitespace-pre-wrap">{video.Description}</p>
            </div>
            {video.Summary && (
              <div className="mt-4 p-4 bg-gray-100 rounded-lg">
                <h2 className="text-sm font-medium text-gray-700 mb-1">Summary generated by {video.SummaryModel}</h2>
                <p className="text-base">{video.Summary}</p>
              </div>
            )}
          </div>
        </div>
        <div>
          <CommentArea videoId={video.ID} />
        </div>
      </div>
    </div>
  );
}

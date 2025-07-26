import React, { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../api/axios';
import { getAuth } from 'firebase/auth';

interface Comment {
  ID: number;
  UserID: string;
  Message: string;
  CreatedAt: string;
  User: {
    Username: string;
  };
}

export default function CommentArea({videoId}:{videoId:string}) {
  const [ws, setWs] = useState<WebSocket|null>(null);
  const [msg, setMsg] = useState('');
  const auth = getAuth();
  const queryClient = useQueryClient();

  const {data:initial} = useQuery<Comment[]>({ 
    queryKey:['comments',videoId], 
    queryFn:()=>api.get(`/v1/videos/${videoId}/comments`).then(r=>r.data)
  });

  const [comments,setComments] = useState<Comment[]>([]);
  useEffect(()=>{ if(initial) setComments(initial); }, [initial]);

  useEffect(()=>{
    const openSocket = async () => {
      const user = auth.currentUser;
      if (!user) return;
      const token = await user.getIdToken();
      const socket = new WebSocket(`${import.meta.env.VITE_WS_URL}/v1/ws/comments?vid=${videoId}&token=${token}`);
      
      socket.onmessage = e => {
        const c:Comment = JSON.parse(e.data);
        setComments(prev => {
          if (prev.find(pc => pc.ID === c.ID)) return prev;
          return [c, ...prev]
        });
      };
      setWs(socket);
    }
    openSocket();

    return () => {
      ws?.close();
    };
  },[videoId, auth.currentUser]);

  const mutation = useMutation({
    mutationFn: (newComment: { video_id: string; message: string }) => {
      return api.post('/v1/comments', newComment);
    },
    onSuccess: () => {
      setMsg('');
      queryClient.invalidateQueries({ queryKey: ['comments', videoId] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate({ video_id: videoId, message: msg });
  };

  return (
    <div className="p-4 bg-gray-100 rounded-lg">
      <h2 className="text-lg font-bold mb-4">Comments</h2>
      <form onSubmit={handleSubmit} className="mb-4">
        <textarea
          value={msg}
          onChange={(e) => setMsg(e.target.value)}
          className="w-full p-2 border rounded-lg"
          placeholder="Add a comment..."
        />
        <button type="submit" className="mt-2 px-4 py-2 bg-blue-500 text-white rounded-lg">Comment</button>
      </form>
      <div>
        {comments.map(comment => (
          <div key={comment.ID} style={{ marginBottom: '1rem', paddingBottom: '0.5rem', borderBottom: '1px solid #e5e7eb' }}>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <p style={{ fontWeight: 'normal' }}>{comment.User?.Username || 'User'}</p>
              <p style={{ color: '#6b7280', fontSize: '0.875rem', marginLeft: '1rem' }}>{new Date(comment.CreatedAt).toLocaleString([], {year: 'numeric', month: 'numeric', day: 'numeric', hour: '2-digit', minute:'2-digit'})}</p>
            </div>
            <p style={{ marginTop: '0' }}>{comment.Message}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

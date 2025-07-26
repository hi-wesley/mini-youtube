import React, { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
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

  // initial fetch (optional)
  const {data:initial} = useQuery<Comment[]>({ 
    queryKey:['comments',videoId], 
    queryFn:()=>api.get(`/v1/videos/${videoId}/comments`).then(r=>r.data)
  });

  const [comments,setComments] = useState<Comment[]>([]);
  useEffect(()=>{ if(initial) setComments(initial); }, [initial]);

  useEffect(()=>{
    const openSocket = async () => {
      console.log('CommentArea: creating WebSocket');
      const user = auth.currentUser;
      if (!user) return;
      const token = await user.getIdToken();
      const socket = new WebSocket(`${import.meta.env.VITE_WS_URL}/v1/ws/comments?vid=${videoId}&token=${token}`);
      
      socket.onopen = () => console.log('CommentArea: WebSocket opened');
      socket.onclose = () => console.log('CommentArea: WebSocket closed');
      socket.onerror = (e) => console.error('CommentArea: WebSocket error:', e);

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
      console.log('CommentArea: closing WebSocket');
      ws?.close();
    };
  },[videoId, auth.currentUser]);

  const mut = useMutation({
    mutationFn:(body:{video_id:string,message:string})=>api.post('/v1/comments',body),
    onSuccess:()=>setMsg('')
  });

  return (
    <div>
      <h4 className="font-semibold mb-2">Comments</h4>
      <form onSubmit={e=>{e.preventDefault(); mut.mutate({video_id:videoId,message:msg});}}>
        <input value={msg} onChange={e=>setMsg(e.target.value)} className="border px-2 py-1 w-full" placeholder="Add a comment..." />
      </form>
      <ul className="my-3 space-y-2">
        {comments.map(c=>(
          <li key={c.ID} className="border-b pb-1">
            <span className="font-medium">{c.User?.Username || 'User'}</span> {c.Message}
          </li>
        ))}
      </ul>
    </div>
  );
}
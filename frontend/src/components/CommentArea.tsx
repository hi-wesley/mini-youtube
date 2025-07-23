import React, { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import api from '../api/axios';
import { getAuth } from 'firebase/auth';

interface Comment {
  id: number;
  user_id: string;
  message: string;
  created_at: string;
}

export default function CommentArea({videoId}:{videoId:string}) {
  const [ws, setWs] = useState<WebSocket|null>(null);
  const [msg, setMsg] = useState('');
  const auth = getAuth();

  // initial fetch (optional)
  const {data:initial} = useQuery<Comment[]>({
    queryKey:['comments',videoId],
    queryFn:()=>api.get(`/v1/videos/${videoId}/comments`).then(r=>r.data),
  });

  const [comments,setComments] = useState<Comment[]>([]);
  useEffect(()=>{ if(initial) setComments(initial); }, [initial]);

  // open WS once
  const initialized = useRef(false);
  useEffect(()=>{
    if(initialized.current) return;
    const socket = new WebSocket(`${import.meta.env.VITE_WS_URL}/v1/ws/comments?vid=${videoId}`);
    socket.onmessage = e => {
      const c:Comment = JSON.parse(e.data);
      setComments(prev=>[c, ...prev]);
    };
    setWs(socket);
    initialized.current = true;
    return ()=>socket.close();
  },[videoId]);

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
          <li key={c.id} className="border-b pb-1">
            <span className="font-medium">{c.user_id.slice(0,6)}</span> {c.message}
          </li>
        ))}
      </ul>
    </div>
  );
}

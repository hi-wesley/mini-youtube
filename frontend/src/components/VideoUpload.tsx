import React, { useState } from 'react';
import api from '../api/axios';
import { Link, useNavigate } from 'react-router-dom';

export default function VideoUpload() {
  const [file, setFile] = useState<File|null>(null);
  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const nav = useNavigate();

  const onSelect = (e:React.ChangeEvent<HTMLInputElement>)=>{
    const f = e.target.files?.[0];
    if(!f) return;
    if(f.size > 500*1024*1024) return alert('Max 500 MB');
    if(!['video/mp4','video/quicktime','video/x-matroska'].includes(f.type))
      return alert('Unsupported format');
    setFile(f);
  };

  const onSubmit = async(e:React.FormEvent)=>{
    e.preventDefault();
    if(!file) return alert('choose a file');
    const fd = new FormData();
    fd.append('video', file);
    fd.append('title', title);
    fd.append('description', desc);
    const {data} = await api.post('/v1/videos', fd, {headers:{'Content-Type':'multipart/form-data'}});
    nav(`/watch/${data.id}`);
  };

  return (
    <div className="max-w-lg mx-auto p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Upload video</h2>
        <Link to="/" className="text-blue-600">Go to Main Page</Link>
      </div>
      <form onSubmit={onSubmit} className="space-y-3">
        <input type="file" onChange={onSelect} accept="video/*"/>
        <input className="border w-full p-1" placeholder="Title" value={title} onChange={e=>setTitle(e.target.value)}/>
        <textarea className="border w-full p-1" placeholder="Description" value={desc} onChange={e=>setDesc(e.target.value)}/>
        <button className="bg-blue-600 text-white px-3 py-1 rounded" disabled={!file}>Upload</button>
      </form>
    </div>
  );
}
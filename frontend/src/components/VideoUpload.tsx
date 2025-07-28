import React, { useState } from 'react';
import api from '../api/axios';
import { Link, useNavigate } from 'react-router-dom';

export default function VideoUpload() {
  const [file, setFile] = useState<File|null>(null);
  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const nav = useNavigate();

  const [error, setError] = useState<string | null>(null);

  const onSelect = (e:React.ChangeEvent<HTMLInputElement>)=>{
    const f = e.target.files?.[0];
    if(!f) return;
    setError(null);
    if(f.size > 100*1024*1024) {
      setError('File size cannot exceed 100 MB.');
      setFile(null);
      return;
    }
    if(!['video/mp4','video/quicktime','video/x-matroska'].includes(f.type)) {
      setError('Unsupported video format.');
      setFile(null);
      return;
    }
    setFile(f);
  };

  const onSubmit = async(e:React.FormEvent)=>{
    e.preventDefault();
    if(!file) return alert('Choose a file');
    if (!title.trim()) return alert('Please provide a video title.');
    if (!desc.trim()) return alert('Please provide a video description.');

    setIsUploading(true);
    const fd = new FormData();
    fd.append('video', file);
    fd.append('title', title);
    fd.append('description', desc);
    try {
      await api.post('/v1/videos', fd, {headers:{'Content-Type':'multipart/form-data'}});
      nav('/');
    } catch (error) {
      console.error('Error uploading video:', error);
      alert('Error uploading video. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto p-4 rounded-lg mt-16" style={{background: 'rgba(255, 255, 255, 0.15)', backdropFilter: 'blur(10px)', WebkitBackdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.2)', boxShadow: '0 8px 32px 0 rgba(59, 130, 246, 0.37)'}}>
      <div className="flex justify-center items-center mb-4">
        <h2 className="text-xl font-bold text-black">Upload a Video</h2>
      </div>
      {error && <p className="text-red-500 text-center mb-4">{error}</p>}
      <form onSubmit={onSubmit} className="space-y-3">
        <label className="flex items-center justify-center gap-2 px-4 h-10 border border-gray-300 rounded-full bg-gray-100 text-black text-sm font-medium cursor-pointer transition-colors hover:bg-gray-200 hover:border-gray-400">
          <span>{file ? file.name : 'Select file...'}</span>
          <input type="file" onChange={onSelect} accept="video/*" className="hidden"/>
        </label>
        <input className="w-full p-2 border rounded-lg" placeholder="Title" value={title} onChange={e=>setTitle(e.target.value)}/>
        <textarea className="w-full p-2 border rounded-lg" placeholder="Description" value={desc} onChange={e=>setDesc(e.target.value)}/>
        <div className="flex justify-center">
          <button className={`px-4 py-2 rounded-lg transition-colors ${file ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-500'}`} disabled={!file || isUploading}>
            {isUploading ? 'Uploading...' : 'Upload'}
          </button>
        </div>
      </form>
    </div>
  );
}
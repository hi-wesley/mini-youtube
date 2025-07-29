// This file defines the component for the video upload page.
// It handles the new direct-to-GCS upload flow, including getting a signed URL,
// uploading the file, and finalizing the process.
import React, { useState } from 'react';
import api from '../api/axios';
import { useNavigate } from 'react-router-dom';
import axios from 'axios'; // Import axios for the direct GCS upload

export default function VideoUpload() {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const [uploadStatus, setUploadStatus] = useState('');
  const [error, setError] = useState<string | null>(null);
  const nav = useNavigate();

  const onSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (!f) return;
    setError(null);
    // Keep the 100MB limit as requested
    if (f.size > 100 * 1024 * 1024) {
      setError('File size cannot exceed 100 MB.');
      setFile(null);
      return;
    }
    if (!['video/mp4', 'video/quicktime', 'video/x-matroska'].includes(f.type)) {
      setError('Unsupported video format. Please use MP4, MOV, or MKV.');
      setFile(null);
      return;
    }
    setFile(f);
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setError('Please select a file to upload.');
      return;
    }
    if (!title.trim()) {
      setError('Please provide a video title.');
      return;
    }
    if (!desc.trim()) {
      setError('Please provide a video description.');
      return;
    }

    setIsUploading(true);
    setError(null);

    try {
      // Step 1: Get the signed URL from our backend
      setUploadStatus('Initializing upload...');
      const initiateResponse = await api.post('/v1/videos/initiate-upload', {
        fileName: file.name,
        fileType: file.type,
      });

      const { uploadUrl, objectName } = initiateResponse.data;

      // Step 2: Upload the file directly to Google Cloud Storage
      setUploadStatus('Uploading video...');
      await axios.put(uploadUrl, file, {
        headers: {
          'Content-Type': file.type,
        },
      });

      // Step 3: Finalize the upload with our backend
      setUploadStatus('Finalizing...');
      await api.post('/v1/videos/finalize-upload', {
        objectName: objectName,
        title: title,
        description: desc,
      });

      setUploadStatus('Upload complete!');
      nav('/'); // Navigate to homepage on success

    } catch (err) {
      console.error('Upload failed:', err);
      let errorMessage = 'An unknown error occurred during upload.';
      if (axios.isAxiosError(err) && err.response) {
        // Handle errors from our API or GCS
        if (err.response.status === 403) {
          errorMessage = 'Permission denied. Could not upload to storage.';
        } else if (err.response.data && err.response.data.error) {
          errorMessage = `Upload failed: ${err.response.data.error}`;
        } else {
          errorMessage = `An error occurred: ${err.message}`;
        }
      }
      setError(errorMessage);
      setUploadStatus('Upload failed.');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto p-4 rounded-lg mt-16" style={{ background: 'rgba(255, 255, 255, 0.15)', backdropFilter: 'blur(10px)', WebkitBackdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.2)', boxShadow: '0 8px 32px 0 rgba(59, 130, 246, 0.37)' }}>
      <div className="flex justify-center items-center mb-4">
        <h2 className="text-xl font-bold text-black">Upload a Video</h2>
      </div>
      {error && <p className="text-red-500 text-center mb-4">{error}</p>}
      <form onSubmit={onSubmit} className="space-y-3">
        <label className="flex items-center justify-center gap-2 px-4 h-10 border border-gray-300 rounded-full bg-gray-100 text-black text-sm font-medium cursor-pointer transition-colors hover:bg-gray-200 hover:border-gray-400">
          <span>{file ? file.name : 'Select file...'}</span>
          <input type="file" onChange={onSelect} accept="video/mp4,video/quicktime,video/x-matroska" className="hidden" />
        </label>
        <input className="w-full p-2 border rounded-lg" placeholder="Title" value={title} onChange={e => setTitle(e.target.value)} />
        <textarea className="w-full p-2 border rounded-lg" placeholder="Description" value={desc} onChange={e => setDesc(e.target.value)} />
        <div className="flex flex-col items-center">
          <button className={`px-4 py-2 rounded-lg transition-colors ${file ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-500'}`} disabled={!file || isUploading}>
            {isUploading ? 'Uploading...' : 'Upload'}
          </button>
          {isUploading && <p className="text-sm text-gray-600 mt-2">{uploadStatus}</p>}
        </div>
      </form>
    </div>
  );
}

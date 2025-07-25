import React from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '../api/axios';

interface User {
  ID: string;
  Email: string;
  Username: string;
}

export default function ProfilePage() {
  const { data: user, isLoading, error } = useQuery<User>({ 
    queryKey: ['profile'], 
    queryFn: () => api.get('/v1/profile').then(res => res.data)
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>An error occurred: {error.message}</div>;

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold">Profile</h1>
      <p><strong>ID:</strong> {user?.ID}</p>
      <p><strong>Email:</strong> {user?.Email}</p>
      <p><strong>Username:</strong> {user?.Username}</p>
    </div>
  );
}

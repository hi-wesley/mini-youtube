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
      <div className="max-w-lg mx-auto p-4 rounded-lg mt-16" style={{background: 'rgba(255, 255, 255, 0.15)', backdropFilter: 'blur(10px)', WebkitBackdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.2)', boxShadow: '0 8px 32px 0 rgba(59, 130, 246, 0.37)'}}>
        <h2 className="text-xl font-bold text-black text-center">Profile</h2>
        <p className="text-black"><strong>Username:</strong> {user?.Username}</p>
        <p className="text-black"><strong>Email:</strong> {user?.Email}</p>
      </div>
    </div>
  );
}

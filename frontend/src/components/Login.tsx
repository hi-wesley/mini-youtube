import React, { useState } from 'react';
import { auth, signInWithEmailAndPassword, createUserWithEmailAndPassword } from './AuthProvider';
import api from '../api/axios';
import Header from './Header';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    
    try {
      if (isRegister) {
        // Check username availability first, before creating Firebase user
        try {
          await api.post('/v1/auth/check-username', { username });
        } catch (error: any) {
          if (error.response?.status === 409) {
            setError('Username already taken. Please choose a different username.');
            return;
          }
          // If it's not a username conflict, proceed with registration
        }
        
        const userCredential = await createUserWithEmailAndPassword(auth, email, password);
        const token = await userCredential.user.getIdToken();
        try {
          await api.post('/v1/auth/register', { username }, { headers: { Authorization: `Bearer ${token}` } });
        } catch (error: any) {
          // This catch block handles errors from your backend API (e.g., username taken)
          if (error.response?.data?.error) {
            setError(error.response.data.error);
          } else {
            setError('An unexpected error occurred during registration.');
          }
          // IMPORTANT: We need to delete the user from Firebase since the backend registration failed
          await userCredential.user.delete();
          return; // Stop execution to stay on the registration page
        }
      } else {
        await signInWithEmailAndPassword(auth, email, password);
      }
    } catch (error: any) {
      // This catch block handles errors from Firebase (e.g., email already in use, weak password)
      if (error.code === 'auth/email-already-in-use') {
        setError('Email already in use.');
      } else if (error.code === 'auth/invalid-email') {
        setError('Invalid email address.');
      } else if (error.code === 'auth/weak-password') {
        setError('Password should be at least 6 characters.');
      } else if (error.code === 'auth/user-not-found' || error.code === 'auth/wrong-password') {
        setError('Invalid email or password.');
      } else {
        setError('An unexpected error occurred. Please try again.');
      }
      console.error(error);
    }
  };

  return (
    <>
    <Header />
    <div className="max-w-lg mx-auto p-4 rounded-lg mt-16" style={{background: 'rgba(255, 255, 255, 0.15)', backdropFilter: 'blur(10px)', WebkitBackdropFilter: 'blur(10px)', border: '1px solid rgba(255, 255, 255, 0.2)', boxShadow: '0 8px 32px 0 rgba(59, 130, 246, 0.37)'}}>
      <div className="flex justify-center items-center mb-4">
        <h2 className="text-xl font-bold text-black">{isRegister ? 'Register' : 'Sign in'}</h2>
      </div>
      {error && <p className="text-red-500 text-center mb-4">{error}</p>}
      <form onSubmit={handleSubmit} className="space-y-3">
        {isRegister && (
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Username"
            className="w-full p-2 border rounded-lg"
          />
        )}
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
          className="w-full p-2 border rounded-lg"
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password"
          className="w-full p-2 border rounded-lg"
        />
        <div className="flex justify-center">
          <button type="submit" className={`px-4 py-2 rounded-lg text-white transition-colors ${email && password ? 'bg-blue-500' : 'bg-gray-300'}`} disabled={!email || !password}>
            {isRegister ? 'Register' : 'Sign in'}
          </button>
        </div>
      </form>
      <div className="flex justify-center mt-4">
        <button onClick={() => setIsRegister(!isRegister)} className="text-sm text-blue-600">
          {isRegister ? 'Already have an account? Login' : 'Need an account? Register'}
        </button>
      </div>
    </div>
    </>
  );
}


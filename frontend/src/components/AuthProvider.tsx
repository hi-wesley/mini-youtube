// This file handles all Firebase authentication logic.
// It initializes the connection to Firebase, creates a "context" to share
// the user's login status across the entire app, and exports functions
// for signing in and creating new users.
import { initializeApp } from 'firebase/app';
import { getAuth, onAuthStateChanged, signInWithEmailAndPassword,
         createUserWithEmailAndPassword, User } from 'firebase/auth';
import React, { createContext, useEffect, useState } from 'react';

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FB_API_KEY,
  authDomain: import.meta.env.VITE_FB_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FB_PROJECT_ID,
};

initializeApp(firebaseConfig);
export const auth = getAuth();

export const AuthCtx = createContext<{user:User|null, loading:boolean} | null>(null);

export default function AuthProvider({children}:{children:React.ReactNode}) {
  const [user, setUser] = useState<User|null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(()=>{
    const unsub = onAuthStateChanged(auth, u=>{
      setUser(u);
      setLoading(false);
    });
    return unsub;
  }, []);
  return <AuthCtx.Provider value={{user, loading}}>{children}</AuthCtx.Provider>;
}

export { signInWithEmailAndPassword, createUserWithEmailAndPassword };

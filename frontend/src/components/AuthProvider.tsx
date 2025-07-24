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

export const AuthCtx = createContext<{user:User|null} | null>(null);

export default function AuthProvider({children}:{children:React.ReactNode}) {
  const [user, setUser] = useState<User|null>(null);
  useEffect(()=>onAuthStateChanged(auth, setUser), []);
  return <AuthCtx.Provider value={{user}}>{children}</AuthCtx.Provider>;
}

export { signInWithEmailAndPassword, createUserWithEmailAndPassword };

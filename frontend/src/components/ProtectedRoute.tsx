import React, { useContext } from 'react';
import { Navigate } from 'react-router-dom';
import { AuthCtx } from './AuthProvider';

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const auth = useContext(AuthCtx);

  if (auth === null) {
    // This can happen while the auth state is initializing.
    // You might want to show a loading spinner here.
    return <div>Loading...</div>;
  }

  if (!auth.user) {
    return <Navigate to="/login" />;
  }

  return <>{children}</>;
}

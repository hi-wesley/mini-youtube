import React, { useContext } from 'react';
import { Navigate } from 'react-router-dom';
import { AuthCtx } from './AuthProvider';

export default function RedirectIfAuth({ children }: { children: React.ReactNode }) {
  const auth = useContext(AuthCtx);

  if (auth === null) {
    return <div>Loading...</div>;
  }

  if (auth.user) {
    return <Navigate to="/" />;
  }

  return <>{children}</>;
}

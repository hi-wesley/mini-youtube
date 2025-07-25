import React, { useContext } from 'react';
import { Link } from 'react-router-dom';
import { AuthCtx } from './AuthProvider';

export default function Header() {
  const auth = useContext(AuthCtx);

  return (
    <header className="bg-white shadow-md">
      <nav className="container mx-auto px-4 py-2 flex justify-between items-center">
        <Link to="/" className="text-xl font-bold">Mini YouTube</Link>
        <div>
          {auth?.user ? (
            <Link to="/profile" className="text-blue-600">Profile</Link>
          ) : (
            <Link to="/login" className="text-blue-600">Login</Link>
          )}
        </div>
      </nav>
    </header>
  );
}

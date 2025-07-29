// This component renders the main navigation header that appears at the top of every page.
// It includes the site logo, a link to the upload page, and dynamic content
// that shows either a "Sign in" button or a user profile dropdown menu
// depending on the user's login status.
import React, { useContext, useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { AuthCtx, auth } from './AuthProvider';
import api from '../api/axios';

const YouTubeIcon = () => (
  <svg width="28" height="20" viewBox="0 0 28 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path 
      d="M27.44 3.12c-.32-1.2-1.27-2.15-2.47-2.47C22.8 0 14 0 14 0S5.2 0 3.03.65c-1.2.32-2.15 1.27-2.47 2.47C0 5.3 0 10 0 10s0 4.7.56 6.88c.32 1.2 1.27 2.15 2.47 2.47C5.2 20 14 20 14 20s8.8 0 10.97-.65c1.2-.32 2.15-1.27 2.47-2.47C28 14.7 28 10 28 10s0-4.7-.56-6.88z" 
      fill="#FF0000"
    />
    <path 
      d="M11.2 14.4V5.6l7.28 4.4-7.28 4.4z" 
      fill="white"
    />
  </svg>
);

const HoverBorderGradient = ({ children, className = "" }: { children: React.ReactNode, className?: string }) => (
  <div className={`relative group ${className}`}>
    <div className="absolute -inset-px bg-gradient-to-r from-blue-300 to-blue-500 rounded-full opacity-0 group-hover:opacity-100 transition duration-300 blur-[1px]"></div>
    <div className="relative bg-white hover:bg-gray-50 rounded-full px-3 py-1 transition duration-300 border border-gray-200">
      {children}
    </div>
  </div>
);

interface User {
  ID: string;
  Email: string;
  Username: string;
}

export default function Header() {
  const authContext = useContext(AuthCtx);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const { data: user } = useQuery<User>({ 
    queryKey: ['profile'], 
    queryFn: () => api.get('/v1/profile').then(res => res.data),
    enabled: !!authContext?.user // Only fetch when user is authenticated
  });

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSignOut = () => {
    auth.signOut();
  };

  const handleCreateClick = (e: React.MouseEvent) => {
    if (!authContext?.user) {
      e.preventDefault();
      alert('Sign in to upload videos');
    }
  };

  return (
    <header className="bg-white shadow-md">
      <nav className="container mx-auto px-4 py-2 flex justify-between items-center">
        <Link to="/" className="flex items-center gap-2 text-xl font-bold" style={{ marginLeft: '0vw' }}>
          <YouTubeIcon />
          Mini YouTube
        </Link>
        <div className="flex items-center gap-4">
          <HoverBorderGradient>
            <Link to="/upload" onClick={handleCreateClick} className="text-blue-600 no-underline">
              + Create
            </Link>
          </HoverBorderGradient>
          {authContext?.user ? (
            <div className="relative" ref={dropdownRef}>
              <HoverBorderGradient>
                <button onClick={() => setDropdownOpen(!dropdownOpen)} className="text-blue-600 no-underline focus:outline-none bg-transparent border-none">
                  {user?.Username || 'User'}
                </button>
              </HoverBorderGradient>
              {dropdownOpen && (
                <div className="absolute right-0 mt-2 py-2 w-48 bg-white rounded-md shadow-xl z-20">
                  <Link to="/profile" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Profile</Link>
                  <button onClick={handleSignOut} className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Sign out</button>
                </div>
              )}
            </div>
          ) : (
            <HoverBorderGradient>
              <Link to="/login" className="text-blue-600 no-underline">
                Sign in
              </Link>
            </HoverBorderGradient>
          )}
        </div>
      </nav>
    </header>
  );
}

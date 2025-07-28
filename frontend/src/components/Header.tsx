import React, { useContext } from 'react';
import { Link } from 'react-router-dom';
import { AuthCtx } from './AuthProvider';

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

const ProfileIcon = () => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 66 65"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="h-3 w-3 text-current"
  >
    <path
      d="M8 8.05571C8 8.05571 54.9009 18.1782 57.8687 30.062C60.8365 41.9458 9.05432 57.4696 9.05432 57.4696"
      stroke="currentColor"
      strokeWidth="15"
      strokeMiterlimit="3.86874"
      strokeLinecap="round"
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

export default function Header() {
  const auth = useContext(AuthCtx);

  return (
    <header className="bg-white shadow-md">
      <nav className="container mx-auto px-4 py-2 flex justify-between items-center">
        <Link to="/" className="flex items-center gap-2 text-xl font-bold" style={{ marginLeft: '0vw' }}>
          <YouTubeIcon />
          Mini YouTube
        </Link>
        <div className="flex items-center gap-4">
          {auth?.user ? (
            <>
              <HoverBorderGradient>
                <Link to="/upload" className="text-blue-600 no-underline">
                  + Create
                </Link>
              </HoverBorderGradient>
              <HoverBorderGradient>
                <Link to="/profile" className="text-blue-600 no-underline">
                  Profile
                </Link>
              </HoverBorderGradient>
            </>
          ) : (
            <Link to="/login" className="text-blue-600">Login</Link>
          )}
        </div>
      </nav>
    </header>
  );
}

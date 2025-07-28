import React from 'react';
import Header from './Header';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{backgroundImage: 'url(https://source.unsplash.com/random/1600x900?nature,water)', backgroundSize: 'cover', backgroundPosition: 'center', backgroundRepeat: 'no-repeat', minHeight: '100vh'}}>
      <Header />
      <main>
        {children}
      </main>
    </div>
  );
}

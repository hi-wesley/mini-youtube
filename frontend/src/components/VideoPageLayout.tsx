import React from 'react';
import Header from './Header';

export default function VideoPageLayout({ children }: { children: React.ReactNode }) {
  return (
    <div>
      <Header />
      {children}
    </div>
  );
}

import React from 'react';

export default function VideoPlayer({ src }: { src: string }) {
  return (
    <video controls src={src} className="w-full rounded-lg">
      Your browser does not support the video tag.
    </video>
  );
}

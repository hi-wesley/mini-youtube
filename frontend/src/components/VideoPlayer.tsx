import React, { useEffect, useRef } from 'react';

export default function VideoPlayer({ src, autoPlay }: { src: string, autoPlay?: boolean }) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (autoPlay && videoRef.current) {
      videoRef.current.play().catch(error => {
        console.error("Autoplay was prevented: ", error);
      });
    }
  }, [src, autoPlay]);

  return (
    <video ref={videoRef} controls src={src} className="w-full rounded-lg" playsInline>
      Your browser does not support the video tag.
    </video>
  );
}

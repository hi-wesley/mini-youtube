import { useEffect, useRef } from 'react';

export default function VideoPlayer({ src, autoPlay }: { src: string, autoPlay?: boolean }) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (autoPlay && videoRef.current) {
      const playVideo = async () => {
        try {
          // Try to play with sound first
          if (videoRef.current) {
            await videoRef.current.play();
          }
        } catch (error) {
          // If it fails, it's likely due to autoplay restrictions.
          // Mute the video and try playing again.
          console.log("Autoplay with sound failed, falling back to muted autoplay.", error);
          if (videoRef.current) {
            videoRef.current.muted = true;
            await videoRef.current.play();
          }
        }
      };
      playVideo();
    }
  }, [src, autoPlay]);

  return (
    // The `muted` attribute is removed from here to allow the initial attempt to play with sound.
    <video ref={videoRef} controls src={src} className="w-full rounded-lg" playsInline>
      Your browser does not support the video tag.
    </video>
  );
}

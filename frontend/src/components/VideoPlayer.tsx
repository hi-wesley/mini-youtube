import Hls from 'hls.js';

export default function VideoPlayer({src}:{src:string}) {
  const ref = useRef<HTMLVideoElement>(null);
  useEffect(()=>{
    if (Hls.isSupported() && ref.current) {
      const hls = new Hls();
      hls.loadSource(src);
      hls.attachMedia(ref.current);
    }
  },[src]);
  return <video controls ref={ref} className="w-full rounded-lg"/>;
}

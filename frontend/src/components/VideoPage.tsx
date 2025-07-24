import { useParams } from 'react-router-dom';

import VideoPlayer from './VideoPlayer';
import CommentArea from './CommentArea';

export default function VideoPage() {
  const { id } = useParams();

  if (!id) {
    return <div>Video not found</div>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'row' }}>
      <div style={{ flex: 3 }}>
        <VideoPlayer videoId={id} />
      </div>
      <div style={{ flex: 1 }}>
        <CommentArea videoId={id} />
      </div>
    </div>
  );
}

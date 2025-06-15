import React, { useEffect, useRef } from 'react';
import shaka from 'shaka-player/dist/shaka-player.compiled';

export default function AuthVideo({ src, token }) {
  const videoEl   = useRef(null);
  const shakaRef  = useRef(null);

  useEffect(() => {
    // 1. create player
    const player = new shaka.Player(videoEl.current);
    shakaRef.current = player;

    // 2. attach Authorization header to every manifest & segment request
    player.getNetworkingEngine().registerRequestFilter((type, request) => {
      const t = shaka.net.NetworkingEngine.RequestType;
      if (type === t.MANIFEST || type === t.SEGMENT) {
        request.headers['Authorization'] = `${token}`;
      }
    });

    // 3. load the source (MP4, DASH or HLS URL)
    player.load(src).catch(console.error);

    // 4. tidy up on unmount
    return () => player.destroy();
  }, [src, token]);

  return <video ref={videoEl} controls style={{ width: '100%' }} />;
}
import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { GetVideoRequest, ListVideosRequest, Video, VideoServiceClient } from "../proto/videoservice"

export const $videos = atom<Video[]>([])

onMount($videos,() => {
    console.log("videos.ts -> onMount()")
    fetchVideos()
})

const unaryInterceptor: UnaryInterceptor<any, any> = {
    intercept: (request, invoker) => {
      const m = request.getMetadata();
      const token = $authToken.get();
      m["authorization"] = token; //`${$authContext.get().user.token}`;
      return invoker(request);
    },
  };
  
export const videoService = new VideoServiceClient(
    import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, ""),
    {},
    {
        unaryInterceptors: [unaryInterceptor],
    }
);

export const fetchVideos = async () => {
    const response = await videoService.ListVideos(ListVideosRequest.fromObject({
        pageNumber: 0,
        pageSize: 10,
    }),{})

    $videos.set(response.videos)
}

export const fetchVideo = async (id: string) => {
    const response = await videoService.GetVideo(GetVideoRequest.fromObject({
         video_id: id
    }),{})

    return response
}